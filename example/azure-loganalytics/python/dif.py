# ---------------------------------------------------------------
# Minimal Python binding for Turbonomic Data Ingestion Framework
# ---------------------------------------------------------------

import json
import time


def del_none(d):
    """
    Delete keys with the value ``None`` in a dictionary, recursively.
    This alters the input so you may wish to ``copy`` the dict first.
    """
    for key, value in list(d.items()):
        if value is None:
            del d[key]
        elif isinstance(value, dict):
            del_none(value)
    return d  # For convenience


class Topology:
    def __init__(self, version="v1", scope=""):
        self.version = version
        self.updateTime = int(time.time())
        self.scope = scope
        self.topology = []

    def AddEntity(self, entity):
        self.topology.append(entity)

    def ToJSON(self):
        return json.dumps(self, default=lambda o: del_none(o.__dict__))


class DIFEntity:
    def __init__(self, uid, entity_type, name=None):
        self.uniqueId = uid
        self.type = entity_type
        self.matchIdentifiers = None
        self.hostedOn = None
        self.partOf = None
        self.metrics = {}

        if name:
            self.name = name
        else:
            self.name = uid

    def AddMetric(self, metric_type, metric_kind, value, key=None):
        if metric_type not in self.metrics:
            metric_list = [DIFMetricVal()]
            self.metrics[metric_type] = metric_list
        else:
            metric_list = self.metrics[metric_type]
        if len(metric_list) < 1:
            return
        metric = metric_list[0]
        if metric_kind == "average":
            metric.average = value
        elif metric_kind == "capacity":
            metric.capacity = value
        if key:
            metric.key = key
        return self

    def Matching(self, matching_id):
        if not self.matchIdentifiers:
            self.matchIdentifiers = DIFMatchingIdentifiers(matching_id)
        return self

    def HostedOn(self, ip_address, host_type='virtualMachine'):
        if not self.hostedOn:
            self.hostedOn = {'hostType': [host_type], 'ipAddress': ip_address}
        return self

    def PartOf(self, uid, entity):
        if not self.partOf:
            self.partOf = [{'uniqueId': uid, 'entity': entity}]
        return self



class DIFMatchingIdentifiers:
    def __init__(self, ip_address):
        self.ipAddress = ip_address


class DIFMetricVal:
    def __init__(self):
        self.average = None
        self.capacity = None
        self.key = None