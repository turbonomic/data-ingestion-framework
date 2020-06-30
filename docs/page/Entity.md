---
layout: default
title: Entity Object
---

The Entity object describes an individual entity. You can create a stand-alone entity, or 
you can stitch an entity into a larget topology via relationships such as 
`hostedOn`, or `partOf`. An entity can include specific metrics.

An entity you create can have one of two functions in your {{ site.data.vars.Product_Short }} 
supply chain:

* Topology Entity
    
     The entity you create exists in the topology, and displays as a member of the supply chain. 
     For example, assume you want to create a business application, and then create services that are 
     members of that business application. You would crate these as Topology Entities.
* Proxy Entity

    The entity you create serves as a place-holder in the topology data you are 
    creating. You can assign metrics to this entity, or add relationships to it to 
    make it part of some other entity. But you also specify the `matchIdentifiers` 
    relationship to map this entity to an entity that already exists in the environment. The 
    match identifier is the IP address of the entity you are mapping to.
    
    When {{ site.data.vars.Product_Short }} loads the topology, it moves your metric and relationship 
    data from the topology file into the matched entity. The proxy entity does not appear 
    in the {{ site.data.vars.Product_Short }} supply chain, but the metric and relationship 
    declarations do appear in the GUI for the matched entity.
    
The Entity object includes the following properties:


<table class="props">
<tr>
    <td><p>type</p>
    </td>
    <td>
    <p>Required String</p>
    <p>Can be one of:</p>
        {% include entityTypesList.html %}
    </td>
</tr>
<tr>
    <td><p>uniqueId</p></td>
    <td>
    <p>Required String</p>
    <p>A unique identifier for the entity. This identifies the entity among all other 
    instances of that type in your topology. Because the ID must be globally unique, 
    you should use naming conventions that express a namespace for this topology 
    that you’re creating.</p></td>
</tr>
<tr>
    <td><p>name</p></td>
    <td><p>
    <p>Required String</p></p>
    <p>The display name for the entity. This is not required to be unique, but it 
    is usually convenient to give unique display names.</p>
    </td>
</tr>
<tr>
    <td><p>matchIdentifiers</p></td>
    <td><p>
    <p>Optional Object</p></p>
    <p>This object contains the IP address of the entity you want to match, as follows:</p>
    <pre>{
    "ipAddress" : "11.22.33.44"
}</pre>
    <p>If you specify <code>matchIdentifiers</code>, then this entity will 
    be a Proxy entity. You specify the IP address of the entity that you want to 
    modify. Then {{ site.data.vars.Product_Short }}  will apply the metrics or the 
    <code>partOf</code> specification from this proxy entity, onto the matching entity.</p>
    </td>
</tr>
<tr>
    <td><p>hostedOn</p></td>
    
    <td>
       <p>Optional Object</p>
       <p>The entity that is a host, or provider, for this entity. This object contains the 
       following properties:</p>
       <ul>
       <li><p><code>hostEntityType</code> (Required Array of String): An array of possible entity 
       types. The probe starts with the first entity type in the list, and checks for a matching 
       Uuid or IP address. It uses the first one to succeed. This array can contain the 
       strings, <code>container</code> and <code>virtualMachine</code>.</p></li>
       <li><p><code>hostUuid</code> (Optional String): The unique identifier for the host. 
             This can come be 
             the Uuid you have declared for some other entity in your topology, or you can 
             use the {{ site.data.vars.Product_Short }} Rest API to get the Uuid of a specific 
             entity in the environment. If you provide a Uuid, you do not need to provide an IP 
             address.</p>
       </li>
       <li><p><code>ipAddress</code> (Optional String): The IP address of a host entity in 
             your environment.
             If you provide an IP address, you do not need to provide a Uuid for the entity.</p>
       </li>
       </ul>
    </td>
</tr>
<tr>
    <td><p>partOf</p></td>
    <td><p>
    <p>Required Array of Objects</p></p>
    <p>An array of entities that contain this entity. For example, one business application can 
    contain multiple service and transaction entities. The business application would be 
    the <i>container</i> or <i>parent</i> entity. Also note that your entity can be contained 
    by more thn one parent. </p>
    <p>The objects that are members of this array identify the entity type and the Uuid of the 
    parent, as follows:</p>
    <pre>{
    "entityType": "businessApplication",
    "ipAddress" : "11.22.33.44"
}</pre>
        <p>The entity type can be one of:</p>
        {% include entityTypesList.html %}
    </td>
</tr>
<tr>
    <td><p>metrics</p></td>
    <td><p>
    <p>Optional Array of metricsEntry</p></p>
    <p>The list of Metric Entry objects that speify metrics and values for this entity. </p>
    <p>See {% include linkToMetricsEntry.html %}</p>
    </td>
</tr>
<table>






