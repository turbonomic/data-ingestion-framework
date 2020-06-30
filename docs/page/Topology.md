---
layout: default
title: Topology Object
---

The Topology object is an array of entities that you want to add to the 
overall {{ site.data.vars.Product_Short }} topology of managed entities. 
When {{ site.data.vars.Product_Short }} loads your DIF data, it displays 
the entities from your topology object in the supply chain.

The Topology object includes the following properties:

<table class="props">
<tr>
    <td><p>version</p>
    </td>
    <td>
    <p>Required String</p>
    <p>The version of DIF that your data complies with. For this version, <code>V1</code></p>
    </td>
</tr>
<tr>
    <td>
    <p>updateTime</p>
    </td>
    <td>
    <p>Required Integer</p>
    <p>Use this to record the time that you generated the data. You should 
    periodically generate new data to reflect changes to your environment. This 
    value tracks when the changes were made.</p>
    </td>
</tr>
<tr>
    <td>
    <p>scope</p>
    </td>
    <td>
    <p>Required String</p>
    <p>The scope is a unique identifier for the topology you are creating. Note that 
    the DIF probe can load multiple topologies from different data servers. Further, 
    these different topologies can al appect shared entities in th environment. The 
    scope effectively declares a namespace that the probe can use to separate 
    its processing of these different topologies.</p>
    </td>
</tr>
<tr>
    <td class="props"><p>topology</p>
    </td>
    <td>
    <p>Required Array of Entity</p>
    <p>The array of entities that make up your topology.</p>
    <p>See {% include linkToEntity.html %}</p>
    </td>
</tr>
</table>