import logging
import os
from http.server import BaseHTTPRequestHandler, HTTPServer

import requests
from azure.loganalytics import models
from msal import ConfidentialClientApplication
from msrest import Serializer, Deserializer

from dif import Topology, DIFEntity

# The host and port this metric server listens on
host_name = ''
port_number = 8081
# Query to get virtual machine name and its IP address
query_vm = """
Heartbeat | summarize arg_max(TimeGenerated, *) by Computer | project Computer, ComputerIP
"""
# Query to get used memory of a virtual machine
query_memory = """
    Perf
    | where TimeGenerated > ago(10m)
    | where ObjectName == "Memory" and
    (CounterName == "Used Memory MBytes" or // the name used in Linux records
    CounterName == "Committed Bytes") // the name used in Windows records
    | summarize avg(CounterValue) by Computer, CounterName, bin(TimeGenerated, 10m)
    | order by TimeGenerated
    """
# Resource ID
resource_id = "https://api.loganalytics.io/.default"
# Resource endpoint
resource_endpoint = "https://api.loganalytics.io/v1/workspaces/{}/query"


class TopologyHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/metrics":
            try:
                # Create topology with entities and metrics
                t = self.create_topology()
                self.respond({'status': 200, 'content': t.ToJSON()})
            except models.ErrorResponseException as error_response:
                self.respond({'status': error_response.response.status_code,
                              'content': error_response.message})

    def respond(self, opts):
        self.send_response(opts['status'])
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        body = opts['content']
        self.wfile.write(bytes(body, 'UTF-8'))

    def login(self):
        app = self.server.app
        scope = self.server.scope
        login_result = app.acquire_token_silent(scope, account=None)
        if not login_result:
            logging.info("No suitable token exists in cache. Get a new one from AAD.")
            login_result = app.acquire_token_for_client(scopes=scope)
        else:
            logging.info("Use token in cache.")
        if "access_token" not in login_result:
            print(login_result.get("error"))
            print(login_result.get("error_description"))
            return None
        return login_result["access_token"]

    def query(self, workspace, query_string):
        token = self.login()
        if token is None:
            return None
        result = None
        headers = {"Authorization": "Bearer " + token,
                   "Content-Type": "application/json"}
        data = {"query": query_string}
        deserializer = self.server.deserializer
        url = resource_endpoint.format(workspace)
        response = requests.post(url, json=data, headers=headers)
        if response.status_code not in [200]:
            err_response = deserializer('ErrorResponse', response)
            logging.error("Failed to query workspace {}. Error: {}. Reason: {}."
                          .format(workspace, err_response.error.code, err_response.error.message))
        else:
            result = deserializer('QueryResults', response)
        return result

    def create_topology(self):
        host_ip_map = {}
        for workspace in self.server.workspaces:
            query_result = self.query(workspace, query_vm)
            if not query_result:
                continue
            for table in query_result.tables:
                for row in table.rows:
                    host_ip_map[row[0]] = row[1]

        t = Topology()
        host_seen = {}
        for workspace in self.server.workspaces:
            query_result = self.query(workspace, query_memory)
            if not query_result:
                continue
            for table in query_result.tables:
                for row in table.rows:
                    host = row[0]
                    if host in host_seen:
                        # This host is already processed
                        continue
                    if host not in host_ip_map:
                        # There is no IP for this host
                        continue
                    host_seen[host] = True
                    host_ip = host_ip_map[host]
                    metric_name = row[1]
                    avg_mem_used = row[3]
                    if metric_name == "Used Memory MBytes":
                        avg_mem_used *= 1024
                    elif metric_name == "Committed Bytes":
                        avg_mem_used /= 1024
                    t.AddEntity(DIFEntity(host, "virtualMachine").
                                AddMetric("memory", "average", avg_mem_used).
                                Matching(host_ip))
        return t


class TopologyServer(HTTPServer):
    def __init__(self, server_address, handler_class=TopologyHandler):
        super().__init__(server_address, handler_class)
        # Retrieve the IDs and secret to use with ServicePrincipalCredentials
        tenant_id = os.environ.get("AZURE_TENANT_ID")
        client_id = os.environ.get("AZURE_CLIENT_ID")
        client_secret = os.environ.get("AZURE_CLIENT_SECRET")
        if not (tenant_id and client_id and client_secret):
            raise ValueError('You must define AZURE_TENANT_ID, AZURE_CLIENT_ID and '
                             'AZURE_CLIENT_SECRET environment variables.')
        # Authority
        authority = "https://login.microsoftonline.com/" + tenant_id
        # Retrieve the list of Workspace IDs, separated by comma
        workspace_ids = os.environ.get("AZURE_LOG_ANALYTICS_WORKSPACES")
        if not workspace_ids:
            raise ValueError('You must define AZURE_LOG_ANALYTICS_WORKSPACES environment '
                             'variable to specify a list of workspace IDs separated by comma.')
        self.app = ConfidentialClientApplication(client_id,
                                                 authority=authority,
                                                 client_credential=client_secret)
        self.scope = [resource_id]
        self.workspaces = str(workspace_ids).split(",")
        client_models = {k: v for k, v in models.__dict__.items() if isinstance(v, type)}
        self.serializer = Serializer(client_models)
        self.deserializer = Deserializer(client_models)


def main():
    logging.basicConfig(format='%(asctime)s %(message)s', datefmt='%m/%d/%Y %I:%M:%S %p',
                        level=logging.INFO)
    server_class = TopologyServer
    httpd = server_class((host_name, port_number), TopologyHandler)
    logging.info('Server Starts - %s:%s' % (host_name, port_number))
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        pass
    httpd.server_close()
    logging.info('Server Stops - %s:%s' % (host_name, port_number))


if __name__ == '__main__':
    main()
