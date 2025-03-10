# Jira ticket example
Goal: Monitor and trend Jira tickets in Turbonomic for a Jira project

Pre-requisite: Jira API Token and user name

JIRA API to get number of tickets:
https://<Your_Domain>.atlassian.net/rest/api/2/search?jql=project=<Project_Name>&maxResults=0


In this example, we will query JIRA to get the total number of tickets for a project, and push that metrics to an existed Business Application in Turbonomic as a KPI commodity
