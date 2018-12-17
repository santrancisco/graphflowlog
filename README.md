### Example app

You can find the example for this app here. The graph is generated  by running it through a set of randomly generated fake flowlog.

[https://sampleflowlog.surge.sh](https://sampleflowlog.surge.sh)

![screenshot](/example.png?raw=true "Popup screenshot")

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


### To build frontend code:

Unfortunately David did not include a README with a build instruction for the project and I was not familiar with nodejs so it took a little longer than expected but after installing grunt, here are the tasks you need to run to build `_site` folder with all html,css,javascript:

```
cd $GOPATH/src/github.com/santrancisco/graphflowlog/frontend/
npm install 
grunt jade
grunt less
grunt browserify
```

### To generate the data

The static site is used to load a data.json file, parse it and graph it. To populate the data.json file with all the network traffic information, I use this little golang app.

```
go get github.com/santrancisco/graphflowlog/cmd/graphflowlog

```

Set the following AWS credential in your environment variable so you can run the app:

```
export AWS_ACCESS_KEY_ID=AKIAJRNSOMEKEY
export AWS_SECRET_ACCESS_KEY=AKIAJRSECRET
export AWS_REGION=us-west-1
```

Run our app with the following switches to turn on verbose mode and output to the frontend folder:
```
cd $GOPATH/src/github.com/santrancisco/graphflowlog/
graphflowlog -v -b {Your flowlog bucket name} \
            -p {The path you exported flowlog to - where the interfaces are listed} \
            -o ./frontend/_site/ \
            -c /tmp/flowlog
```

Now we should have our static site under ./frontend/_site/. We can use python SimpleHTTPServer to host the site on port 8080 localhost:

```
cd $GOPATH/src/github.com/santrancisco/graphflowlog/frontend/_site
python -m SimpleHTTPServer 8080
```


