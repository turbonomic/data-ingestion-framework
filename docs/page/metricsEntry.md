---
layout: default
title: Metrics Entry Object
---

The Metrics Entry object expresses a single metric, with its values for utilization and 
capacity.  You should be aware of the following points concerning metrics:


* The given metric types are all types that 
  {{ site.data.vars.Product_Short }} supports natively in its supply chain *with the exception of* kpi metrics
  
* The kpi metric is for Key Performance Indicators. You can track any indicator you want, against a capacity 
  that you specify.  Charts will include this metric, and will show the percentage of capacity 
  your entities are utilizing. For example, you could track the number of issues for a 
  given application, and charts would show when the too many issues are currently open. 
  Note that these metrics do not affect {{ site.data.vars.Product_Short }} analysis or the 
  actions it generates.
  
* For native metrics that you track for your entities, charts display the values. However, these 
  metrics do not affect {{ site.data.vars.Product_Short }} analysis or the 
  actions it generates.
  
* Not all metric types make sense for all entity types. For example, dbMem makes sense for 
  databaseServer entities, but not for a businessTransaction entity. For more information, 
  see {% include linkToEntCommodities.html %}

Each metrics entry object is of a specific type. For each type, you express the average, min, 
max, and capacity values in a specific metruc unit. The following table describes each metric type:


<table class="props">
<tr>
    <td><p>kpi</p>
    </td>
    <td>
    <p>Key Performance Indicator</p>
    <p>Metric Unit: <code>count</code></p>
    <p>A custom metric that you can use to track entities by different criteria. For example, 
    you can use kpi to track the number of support tickets that are open for a given 
    application.</p>
    <p>The <code>key</code> that you provide gives the name of the metric that you will 
    see in charts and tables in the {{ site.data.vars.Product_Short }} user interface.</p>
    </td>
</tr>
<tr>
    <td><p>responseTime</p>
    </td>
    <td>
    <p>Response Time</p>
    <p>Metric Unit: <code>ms</code></p>
    <p>A measure of the time it takes for an entity (for example, an application or a service)
    to process a transaction.</p>
    </td>
</tr>
<tr>
    <td><p>transaction</p>
    </td>
    <td>
    <p>Transactions Per Second</p>
    <p>Metric Unit: <code>tps</code></p>
    </td>
</tr>
<tr>
    <td><p>connections</p>
    </td>
    <td>
    <p>Concurrent Connections</p>
    <p>Metric Unit: <code>count</code></p>
    </td>
</tr>
<tr>
    <td><p>heap</p>
    </td>
    <td>
    <p>Amount of Heap in Use</p>
    <p>Metric Unit: <code>mb</code></p>
    </td>
</tr>
<tr>
    <td><p>collectionTime</p>
    </td>
    <td>
    <p>Time Spent on Garbage Collection</p>
    <p>Metric Unit: <code>pct</code></p>
    <p>For a 10 minute period, the percentage of that time that is spent 
    on garbage collection.</p>
    </td>
</tr>
<tr>
    <td><p>cpu</p>
    </td>
    <td>
    <p>Utilization of CPU Capacity</p>
    <p>Metric Unit: <code>mhz</code></p>
    <p>The schema specifies a rawData object for this metric. The rawData is not used at this time.</p>
    </td>
</tr>
<tr>
    <td><p>memory</p>
    </td>
    <td>
    <p>Utilization of Memory</p>
    <p>Metric Unit: <code>mb</code></p>
    <p>The schema specifies a rawData object for this metric. The rawData is not used at this time.</p>
    </td>
</tr>
<tr>
    <td><p>threads</p>
    </td>
    <td>
    <p>Concurrent Threads for the Entity</p>
    <p>Metric Unit: <code>count</code></p>
    <p></p>
    </td>
</tr>
<tr>
    <td><p>cacheHitRate</p>
    </td>
    <td>
    <p>Rate of Data Requests Satisfied by Cache</p>
    <p>Metric Unit: <code></code></p>
    <p>Percentage of data requests that are satisfied by the database cache, where a greater 
    value indicates fewer disk reads for data. The capacity is assumed to be 100% – You do 
    not need to provide a value for capacity.</p>
    </td>
</tr>
<tr>
    <td><p>dbMem</p>
    </td>
    <td>
    <p>Utilization of Database Memory</p>
    <p>Metric Unit: <code>mb</code></p>
    <p>When a database has its own allocated capacity of memory, this tracks the utilization 
    of that capacity.</p>
    </td>
</tr>
</table>