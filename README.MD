#

ApiGateway specifies Attribute "ServiceLabel" with a certain Vaule x
Controller creates an Ingress with Label "ServiceLabel" with Value x
Services that have a Label ServiceLabel with Value x are added to the corresponding ingress
Serice has to specify host and can specify path

Use Cases

New ApiGatesway: create Ingress --> Add Services
Changed ApiGateway, New ServiceLabel: Change Ingress

Delete INgress; If Apigate exists, recreate ingress
UpdateIngress: Chack if anything relavcasnt was changed and change it back to apigateway spec.

New Service: Check INgresses and and add this one service
UpdateService: Check if any relevant attribute has changed and update ingress
DeleteService: Delete Service from ingress

Throw events

AddService in Bulk and as a single Service is created

