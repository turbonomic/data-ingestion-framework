---
layout: default
title: Commodities Bought and Sold by Entities
---

<p>Entities in a {{ site.data.vars.Product_Short }} supply chain exist in 
relationship to each other.  One entity can sell commodities to some entities, 
and it can buy commodities from other entities. A commodity is the resource 
that is expressed by a {% include linkToMetricsEntry.html %}
</p>

There are constraints to these relationships. For example, a businessApplication cannot 
buy dbMem from a businessTransaction entity, but it can buy dbMem from a databaseServer 
entity. 

The following table lists entity types, and the commodities they can buy and sell:


<table>
   <tr>
      <th>Entity Type</th>
      <th>Commodities Sold</th>
      <th>Commodities Bought</th>
   </tr>
   <tr>
      <td>
         <p>businessApplication</p>
      </td>
      <td>
         <ul>
            <li>kpi</li>
         </ul>
      </td>
      <td>
         <ul>
            <li>
               <p>Bought From: 
                  service
               </p>
               <ul>
                  <li>
                     <p>transction</p>
                  </li>
                  <li>
                     <p>responseTime</p>
                  </li>
                  <li>
                     <p>kpi</p>
                  </li>
               </ul>
            </li>
            <li>
               <p>Bought From: 
                  businessTransaction
               </p>
               <ul>
                  <li>
                     <p>transaction</p>
                  </li>
                  <li>
                     <p>responseTime</p>
                  </li>
                  <li>
                     <p>kpi</p>
                  </li>
               </ul>
            </li>
         </ul>
      </td>
   </tr>
   <tr>
      <td>
         <p>businessTransaction</p>
      </td>
      <td>
         <ul>
            <li>transaction</li>
            <li>responseTime</li>
            <li>kpi</li>
         </ul>
      </td>
      <td>
         <ul>
            <li>
               <p>Bought From: 
                  service
               </p>
               <ul>
                  <li>
                     <p>transaction</p>
                  </li>
                  <li>
                     <p>responseTime</p>
                  </li>
                  <li>
                     <p>kpi</p>
                  </li>
               </ul>
            </li>
         </ul>
      </td>
   </tr>
   <tr>
      <td>
         <p>service</p>
      </td>
      <td>
         <ul>
            <li>transaction</li>
            <li>responseTime</li>
            <li>heap</li>
            <li>collectionTime</li>
            <li>threads</li>
            <li>kpi</li>
         </ul>
      </td>
      <td>
         <ul>
            <li>
               <p>Bought From: 
                  application
               </p>
               <ul>
                  <li>
                     <p>transaction</p>
                  </li>
                  <li>
                     <p>responseTime</p>
                  </li>
                  <li>
                     <p>heap</p>
                  </li>
                  <li>
                     <p>collectionTime</p>
                  </li>
                  <li>
                     <p>threads</p>
                  </li>
                  <li>
                     <p>kpi</p>
                  </li>
               </ul>
            </li>
            <li>
               <p>Bought From: 
                  databaseServer
               </p>
               <ul>
                  <li>
                     <p>transaction</p>
                  </li>
                  <li>
                     <p>responseTime</p>
                  </li>
                  <li>
                     <p>kpi</p>
                  </li>
               </ul>
            </li>
         </ul>
      </td>
   </tr>
   <tr>
      <td>
         <p>application</p>
      </td>
      <td>
         <ul>
            <li>transaction</li>
            <li>responseTime</li>
            <li>heap</li>
            <li>collectionTime</li>
            <li>threads</li>
            <li>kpi</li>
         </ul>
      </td>
      <td>
         <ul>
            <li>
               <p>Bought From: 
                  virtualMachine
               </p>
               <ul>
                  <li>
                     <p>cpu</p>
                  </li>
                  <li>
                     <p>memory</p>
                  </li>
               </ul>
            </li>
            <li>
               <p>Bought From: 
                  container
               </p>
               <ul>
                  <li>
                     <p>cpu</p>
                  </li>
                  <li>
                     <p>memory</p>
                  </li>
               </ul>
            </li>
         </ul>
      </td>
   </tr>
   <tr>
      <td>
         <p>databaseServer</p>
      </td>
      <td>
         <ul>
            <li>transaction</li>
            <li>responseTime</li>
            <li>connections</li>
            <li>dbMem</li>
            <li>cacheHitRate</li>
            <li>kpi</li>
         </ul>
      </td>
      <td>
         <ul>
            <li>
               <p>Bought From: 
                  virtualMachine
               </p>
               <ul>
                  <li>
                     <p>cpu</p>
                  </li>
                  <li>
                     <p>memory</p>
                  </li>
               </ul>
            </li>
         </ul>
      </td>
   </tr>
   <tr>
      <td>
         <p>virtualMachine</p>
      </td>
      <td>
         <ul>
            <li>cpu</li>
            <li>memory</li>
         </ul>
      </td>
      <td>
         <ul></ul>
      </td>
   </tr>
   <tr>
      <td>
         <p>container</p>
      </td>
      <td>
         <ul>
            <li>cpu</li>
            <li>memory</li>
         </ul>
      </td>
      <td>
         <ul></ul>
      </td>
   </tr>
</table>


