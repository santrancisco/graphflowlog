### Info

This is a small project i did to map out infrastructure using flow log. The go app is used to download flowlog samples from s3 bucket and export the data of unique connections between machines inside the VPC. This data is then used by the frontend backbonejs app that I modified the code from David Mcclure's humanist [app](https://github.com/davidmcclure/humanist). This app has beautiful presentation with search bar, pop up and impressively fast when i tested it in chrome.

The app was tested only with AWS credentials being set as the usual environment variables. 

The graph has some unique features that are quite different to any other tools out there:

  - All nodes are static and have exact coordinate. These coordinates are pre-calculated base off its IP address
  - Each collumn is a /24 subnet (eg you will find all 10.65.232.0/24 under a same collumn)
  - The Y axis is determined by the last octet of the IP (eg 10.65.232.1 is most likely be at the bottom and .255 will be at the top) 
  - The open ports are used to calculate the color of the node so nodes with the same list of open port will have the same color.

Note :
  - All "open ports" are determined by simply comparing fromport and toport and choose the lesser value so there maybe some wrong report when the listener is on a very high port
  - When zoom out of the graph, it will only display upto 1000 connections and not show the rest.
