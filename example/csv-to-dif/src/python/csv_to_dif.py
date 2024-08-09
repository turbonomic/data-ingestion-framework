import csv
import json
import sys
import logging
import re
import os
from hashlib import sha256
from collections import defaultdict
from http.server import BaseHTTPRequestHandler, HTTPServer
from io import StringIO

import umsg
import boto3
import azure.core.exceptions
import botocore.exceptions
from azure.storage.blob import BlobServiceClient
from turbodif import Topology, DIFEntity


class IpAddressNotFoundError(Exception):
    pass


class ParentEntityNotFoundError(Exception):
    pass


class InvalidParentTypeError(Exception):
    pass


class InvalidMemberTypeError(Exception):
    pass


class InvalidMetricTypeError(Exception):
    pass


class InvalidConfigError(Exception):
    pass


class CsvDownloadError(Exception):
    def __init__(self, message):
        self.status_code = 404
        self.message = message


class UserDefinedApp():
    """Represents user-defined application topology
    """
    acceptable_parent_types = {
        'businessApplication': {''},
        'businessTransaction': {'businessApplication'},
        'service': {'businessApplication', 'businessTransaction'},
        'databaseServer': {'service'},
        'application': {'service'},
        'virtualMachine': {'application', 'databaseServer',
                           'businessApplication'},
        'container': {'application'}
    }

    acceptable_metric_types = {
        'businessApplication': {'kpi'},
        'businessTransaction': {'transaction', 'responseTime', 'kpi'},
        'service': {'transaction', 'responseTime', 'heap',
                    'collectionTime', 'threads', 'kpi'},
        'databaseServer': {'transaction', 'responseTime', 'heap',
                           'collectionTime', 'threads', 'kpi'},
        'application': {'transaction', 'responseTime', 'connections',
                        'dbMem', 'cacheHitRate', 'kpi'},
        'virtualMachine': {'cpu', 'memory'},
        'container': {'cpu', 'memory'}
    }

    def __init__(self, app_name, dif_topology, prefix=''):
        # TODO: Add the prefix for the Business App name
        self.name = app_name
        self.app_prefix = prefix
        self.members = {}
        self.member_metrics = defaultdict(dict)
        self.dif_topology = dif_topology
        self.add_member(self.name, 'businessApplication')

    @staticmethod
    def _process_ips(ips):
        ip_addresses = []
        ip_regex = r"\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b"

        if isinstance(ips, list):
            for ip in ips:
                matches = re.findall(ip_regex, ips)
                ip_addresses.extend(matches)

        if isinstance(ips, str):
            matches = re.findall(ip_regex, ips)
            ip_addresses.extend(matches)

        if not ip_addresses:
            raise IpAddressNotFoundError()

        return ','.join(ip_addresses)

    def _make_dummy_info(self, entity_type, parent_name, parent_type):
        name = f'{self.name}_{entity_type}'
        hashed = sha256()
        hashed.update(name.encode('utf-8'))
        return {'name': name,
                'type': entity_type,
                'uid': hashed.hexdigest(),
                'parent_name': parent_name,
                'parent_type': parent_type}

    def _make_dummy_levels(self):
        umsg.log(f'Creating dummy application and service for {self.name}',
                 level='debug')
        service_name = f'{self.name}_service'
        app_name = f'{self.name}_application'
        dummy_service = self._make_dummy_info('service', self.name,
                                              'businessApplication')
        dummy_app = self._make_dummy_info('application', service_name,
                                          'service')
        self.members[(service_name, 'service')] = dummy_service
        self.members[(app_name, 'application')] = dummy_app

    def is_valid_parent(self, member_type, parent_type):
        try:
            if parent_type in self.acceptable_parent_types[member_type]:
                return True

            else:
                raise InvalidParentTypeError()

        except KeyError:
            raise InvalidMemberTypeError()

    def is_valid_metric(self, member_type, metric_type):
        try:
            if metric_type in self.acceptable_metric_types[member_type]:
                return True

            else:
                raise InvalidMetricTypeError()

        except KeyError:
            raise InvalidMemberTypeError()

    def add_member(self, member_name, member_type, parent_name=None, parent_type=None, member_ip=None):
        if member_type != 'businessApplication':
            try:
                self.is_valid_parent(member_type, parent_type)

            except InvalidParentTypeError as e:
                msg = f'{member_name} does not have a valid parent type.'
                dbug_msg = (f'Member type: [ {member_type} ], provided parent type: [ {parent_type} ], ' +
                            f'valid parent types: [ {", ".join(self.acceptable_parent_types[member_type])} ]')
                umsg.log(msg, level='error')
                umsg.log(dbug_msg, level='debug')
                raise e

        hashed = sha256()
        hashed.update(f'{member_name}_{member_type}'.encode('utf-8'))
        member_info = {'name': member_name,
                       'type': member_type,
                       'uid': hashed.hexdigest(),
                       'parent_name': parent_name,
                       'parent_type': parent_type}

        if (member_name, member_type) in self.members:
            umsg.log(f'Member {member_name} already exists in {self.name} application group',
                     level='warn')

        elif member_type == 'virtualMachine':
            if parent_type == 'businessApplication':
                dbug_msg = (f'VirtualMachine [ {member_name} ] has a parent type of BusinessApplication. ' +
                            f'Attaching to dummy application named [ {self.name}_application ]')
                umsg.log(dbug_msg, level='debug')

                if (f'{self.name}_service', 'service') not in self.members:
                    if parent_name == self.name:
                        self._make_dummy_levels()

                    else:
                        msg = f'Parent for {member_name} with parent type BusinessApplication does not match defined BusinessApplication name'
                        raise ParentEntityNotFoundError(msg)

                member_info.update({'parent_name': f'{self.name}_application',
                                    'parent_type': 'application'})

            try:
                member_info['ip_address'] = self._process_ips(member_ip)

            except IpAddressNotFoundError:
                member_info['ip_address'] = None
                umsg.log(f'No IP address provided for VM named: {member_name}',
                         level='warn')

        self.members[(member_name, member_type)] = member_info

    def add_metrics(self, member_name, member_type, metrics):
        for metric_type, metric_val in metrics.items():
            metric_type, cap_or_used = metric_type.split('_')

            if metric_val:
                try:
                    self.is_valid_metric(member_type, metric_type)

                except InvalidMetricTypeError:
                    msg = f'{member_name} has an invalid metric type, skipping metric named: {metric_type}'
                    dbug_msg = (f'Member type: [ {member_type} ], provided metric type: [ {metric_type} ], ' +
                                f'valid metric types: [ {", ".join(self.acceptable_metric_types[member_type])} ]')
                    umsg.log(msg, level='warning')
                    umsg.log(dbug_msg, level='debug')

                self.member_metrics[(member_name, member_type)][(metric_type, cap_or_used)] = float(metric_val)

    def add_member_to_dif_topo(self, member):
        if member['type'] == 'businessApplication':
            member_name = f"{self.app_prefix}{member['name']}"
        else:
            member_name = member['name']

        dif_entity = DIFEntity(uid=member['uid'],
                               entity_type=member['type'],
                               name=member_name)

        # Parent-child matching
        if member['type'] != 'businessApplication':
            try:
                parent_key = (member['parent_name'],  member['parent_type'])
                parent_entity = self.members[parent_key]
                dif_entity.PartOf(parent_entity['uid'], parent_entity['type'])

            except KeyError:
                raise ParentEntityNotFoundError()

            if member['type'] == 'virtualMachine' and member.get('ip_address'):
                dif_entity.Matching(member['ip_address'])

        self.add_metrics_to_dif_entity(dif_entity)
        self.dif_topology.AddEntity(dif_entity)

    def add_metrics_to_dif_entity(self, dif_entity):
        member_key = (dif_entity.name, dif_entity.type)
        for m_name, m_val in self.member_metrics[member_key].items():
            dif_entity.AddMetric(m_name[0], m_name[1], m_val)

    def create_dif_entities(self):
        for member_info in self.members.values():
            try:
                self.add_member_to_dif_topo(member_info)

            except ParentEntityNotFoundError:
                msg = f"No parent found for entity named: {member_info['name']} with type: {member_info['type']}, skipping"
                umsg.log(msg, level='error')

        return self.dif_topology


class DifCsvReader():
    def __init__(self, filename, csv_location, entity_headers, metric_headers):
        self.filename = filename
        self.entity_headers = entity_headers
        self.metric_headers = metric_headers
        self._check_headers()
        self.process_csv_location(csv_location)

    def _process_entity_headers(self, row):
        entities = {}
        for k, v in self.entity_headers.items():
            try:
                entities[k] = row[v]

            except KeyError:
                umsg.log(f'Incorrect entity field map entry: key: {k}, value: {v}', level='error')
                raise

        return entities

    def _process_metric_headers(self, row):
        metrics = {}
        for k, v in self.metric_headers.items():
            try:
                metrics[k] = row[v]

            except KeyError:
                if v:
                    umsg.log(f'Incorrect metric field map entry: key: {k}, value: {v}', level='error')
                    raise

        return metrics

    def _check_headers(self):
        """Initialize default header dictionaries based on header type"""
        valid_entity_columns = {'app_name', 'entity_name', 'entity_type',
                                'entity_ip', 'parent_name', 'parent_type'}
        valid_metric_names = {'memory', 'cpu', 'heap', 'kpi', 'cacheHitRate',
                              'collectionTime', 'threads', 'responseTime',
                              'dbMem', 'transaction', 'connections'}
        valid_metric_types = {'capacity', 'average', 'max', 'min'}

        if self.entity_headers:
            bad_headers = self.entity_headers.keys() - valid_entity_columns
            if bad_headers:
                msg = f'The following entity field map keys are invalid: {", ".join(bad_headers)}'
                umsg.log(msg, level='error')
                raise InvalidConfigError(msg)

        else:
            umsg.log(f'No CSV entity header mapping provided, using defaults',
                     level='warn')
            self.entity_headers = {x: x for x in valid_entity_columns}

        if self.metric_headers:
            bad_headers = []
            for k in self.metric_headers.keys():
                mname, mtype = k.split('_')
                if mname not in valid_metric_names or mtype not in valid_metric_types:
                    bad_headers.append(k)

            if bad_headers:
                msg = f'The following metric field map keys are invalid: {", ".join(bad_headers)}'
                umsg.log(msg, level='error')
                raise InvalidConfigError(msg)

        else:
            umsg.log(f'No CSV metric header mapping provided, using defaults',
                     level='warn')
            self.metric_headers = {f'{mname}_{mtype}': f'{mname}_{mtype}'
                                   for mname in valid_metric_names
                                   for mtype in valid_metric_types}

    def process_csv_location(self, provider):
        if provider not in {'AZURE', 'AWS', 'FTP'}:
            umsg.log('Value for CSV_LOCATION is invalid. It must be one of: [ AZURE, AWS, FTP ]',
                     level='error')
            raise InvalidConfigError()

        if provider == 'AZURE':
            self.provider = 'AZURE'
            self.connect_str = os.environ['AZURE_CONNECTION_STRING']
            self.container_name = os.environ['AZURE_CONTAINER_NAME']

        if provider == 'AWS':
            self.provider = 'AWS'
            self.access_key_id = os.environ['AWS_ACCESS_KEY_ID']
            self.secret_access_key = os.environ['AWS_SECRET_ACCESS_KEY']
            self.region_name = os.environ['AWS_REGION_NAME']
            self.bucket_name = os.environ['AWS_BUCKET_NAME']

        if provider == 'FTP':
            self.provider = 'FTP'
            self.path = '/opt/turbonomic/data'

    def download_csv_data(self):
        umsg.log(f'Downloading CSV data from {self.provider}')
        try:
            if self.provider == 'AZURE':
                service_client = BlobServiceClient.from_connection_string(self.connect_str)
                blob_client = service_client.get_blob_client(container=self.container_name,
                                                             blob=self.filename)
                file_data = blob_client.download_blob().readall()
                file = file_data.decode('utf-8-sig')

            if self.provider == 'AWS':
                s3_client = boto3.resource(service_name='s3',
                                           region_name=self.region_name,
                                           aws_access_key_id=self.access_key_id,
                                           aws_secret_access_key=self.secret_access_key)
                try:
                    file_data = s3_client.Object(self.bucket_name, self.filename).get()['Body'].read()
                    file = file_data.decode('utf-8-sig')

                except s3_client.exceptions.NoSuchKey:
                    raise FileNotFoundError

            if self.provider == 'FTP':
                filepath = os.path.join(self.path, self.filename)
                with open(filepath, 'r', encoding='utf-8-sig') as fp:
                    file = fp.read()

        except (botocore.exceptions.ClientError,
                botocore.exceptions.InvalidRegionError,
                azure.core.exceptions.HttpResponseError) as e:
            msg = f'Error connecting to cloud provider: {e}'
            umsg.log(msg, level='error')
            raise CsvDownloadError(msg)

        except (azure.core.exceptions.ResourceNotFoundError, FileNotFoundError):
            msg = 'CSV file not found'
            umsg.log(msg, level='error')
            raise CsvDownloadError(msg)

        return StringIO(file)

    def read_csv(self, csv_str_io):
        """Parse CSV StringIO to dict
        Parameters:
            filename - StringIO - IO data from CSV file

        Returns:
            List of dicts, where each dict is a row in the input CSV
        """
        data = []
        csv_data = csv.DictReader(csv_str_io)

        row_count = 1
        for row in csv_data:
            row_count += 1

            if not row[self.entity_headers['app_name']]:
                umsg.log(f'No application defined on row {row_count} of input CSV, skipping',
                         level='warn')
                continue

            try:
                data.append((self._process_entity_headers(row),
                             self._process_metric_headers(row)))

            except KeyError:
                umsg.log(f'Something went wrong on line {row_count} while processing CSV')
                raise

        return data


def MakeHandlerClassFromArgs(csv_filename, csv_location, entity_headers, metric_headers, app_prefix):
    class TopologyHandler(BaseHTTPRequestHandler):
        def __init__(self, *args, **kwargs):
            self.csv_filename = csv_filename
            self.csv_location = csv_location
            self.entity_headers = entity_headers
            self.metric_headers = metric_headers
            self.app_prefix = app_prefix
            super(TopologyHandler, self).__init__(*args, **kwargs)

        def do_GET(self):
            if self.path == '/dif_metrics':
                try:
                    topology = self.create_topology()
                    umsg.log('Serving DIF topology JSON')
                    self.respond({'status': 200,
                                  'content': topology.ToJSON()})

                except CsvDownloadError as e:
                    self.respond({'status': e.status_code,
                                  'content': e.message})

                except Exception as e:
                    self.respond({'status': 500,
                                  'content': f'Something went wrong: {e}'})

        def respond(self, opts):
            self.send_response(opts['status'])
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            body = opts['content']
            self.wfile.write(bytes(body, 'UTF-8'))

        def get_csv_data(self):
            reader = DifCsvReader(filename=self.csv_filename,
                                  csv_location=self.csv_location,
                                  entity_headers=self.entity_headers,
                                  metric_headers=self.metric_headers)

            data = reader.download_csv_data()
            return reader.read_csv(data)

        def create_topology(self):
            data = self.get_csv_data()
            topology = Topology()
            apps = parse_data_into_apps(data, topology, self.app_prefix)
            umsg.log('Building DIF topology...')
            build_dif_topology(apps)

            return topology

    return TopologyHandler


def read_config_file(config_file):
    """Read JSON config file
    Parameters:
        config_file - str - Name of JSON config file
    Output:
        Dict - Dict representation of JSON config file
    """
    with open(config_file, 'r') as fp:
        try:
            return json.loads(fp.read())

        except TypeError:
            umsg.log(f'{config_file} must be JSON format', level='error')


def parse_data_into_apps(csv_data, dif_topology, prefix):
    """Parse input CSV into dictionary of UserDefinedApps
    Parameters:
        csv_data - list - List of dicts from read_csv
        dif_topology - dif.Topology - DIF Topology object
        entity_headers - dict - Optional mapping for CSV entity column names
        metric_headers - dict - Optional mapping for CSV metric column names
    """
    app_dict = {}
    row_count = 1

    umsg.log('Looking for apps and associated VMs...')

    for entity, metric in csv_data:
        row_count += 1
        app_name = entity['app_name']

        if app_name in app_dict.keys():
            app = app_dict[app_name]

        else:
            app = UserDefinedApp(app_name, dif_topology, prefix)
            app_dict[app_name] = app

        try:
            app.add_member(member_name=entity['entity_name'],
                           member_type=entity['entity_type'],
                           parent_name=entity['parent_name'],
                           parent_type=entity['parent_type'],
                           member_ip=entity['entity_ip'])

        except InvalidParentTypeError:
            msg = f'Invalid parent_type found on row {row_count} of input CSV'
            raise InvalidParentTypeError(msg)

        try:
            app.add_metrics(member_name=entity['entity_name'],
                            member_type=entity['entity_type'],
                            metrics=metric)

        except Exception as e:
            raise e

    return app_dict


def build_dif_topology(apps):
    """Add entities to DIF topology by BusinessApplication
    Parameters:
        apps - dict - Mapping of BusinessApplication name to UserDefinedApp object
    """
    for app in apps.values():
        app.create_dif_entities()


def write_topology(topology, output_file):
    """Write DIF topology to JSON file
    Parameters:
        topology - dif.Topology - Full DIF topology with all relevant entities and BusinessApps added
        output_file - str - Output filename for writing
    """
    output = topology.ToJSON()

    with open(output_file, 'w') as fp:
        fp.write(output)


def main(config_file):
    args = read_config_file(config_file)
    host_name = '0.0.0.0'
    port_number = 8081
    umsg.init(level=args.get('LOG_LEVEL', 'INFO'))
    log_file = os.path.join(args.get('LOG_DIR'), args.get('LOG_FILE'))

    if log_file:
        handler = logging.handlers.RotatingFileHandler(log_file,
                                                       mode='a',
                                                       maxBytes=10*1024*1024,
                                                       backupCount=1,
                                                       encoding=None,
                                                       delay=0)
        umsg.add_handler(handler)

    else:
        umsg.add_handler(logging.StreamHandler())

    topology_handler = MakeHandlerClassFromArgs(args['INPUT_CSV_NAME'],
                                                args['CSV_LOCATION'],
                                                args.get('ENTITY_FIELD_MAP'),
                                                args.get('METRIC_FIELD_MAP'),
                                                args.get('APP_PREFIX'))

    httpd = HTTPServer((host_name, port_number), topology_handler)

    umsg.log(f'Server Starts - {host_name}:{port_number}')
    try:
        httpd.serve_forever()

    except KeyboardInterrupt:
        pass

    httpd.server_close()
    umsg.log(f'Server Stops - {host_name}:{port_number}')


if __name__ == '__main__':
    main(sys.argv[1])
