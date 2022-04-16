using an inverted index on the posts

should store questions and answers separately?

can index answers and questions separately, or together. not clear
but together seems better for ranking results

separately seems to be easier than re-indexing

indexing seems to work pretty well? according to lucene benchmark

can index hundreds of GB per hour using some huge machine.

storage seems fairly small as well. can probably expect to leave small footprint if not storing the text data used.

index files are compressed? seems so.

memory required to read and search index seems small. 

we should calculate how many words we expect to see and how much space we expect to use in an inverted index.

we also shouold calculate how much memory is needed and how fast we can get results, join them with items and parse the results.

grows a doubling yearly rate??? may have to handle much higher volumes in the future, and very high peaks.

one option is to? store the index with text and return text with links coming later

the other option is to store at least the views and votes to make use of search?

feel it's better to make some custom search code which take the term queries and joins with view and vote data.


10M questions
50M answers
100-125M reads per year

growth is double per year

each items is about 2-10KB

at 2 KB
1year
10M questions = 20GB data
50M answers = 100GB data
total = 120GB

2year
20M questions = 40GB
100M questions = 200GB data
total = 240

5year
160M questions = 960GB data
800M answers = 1.6TB data
total = 2.6TB data

at 10KB
1year
total = 600GB

3year
total = 1.2TB

5year
total = 13TB

1year
125M reads

2year 250M reads

5year 2B reads per

125 * 2^5 - 125 * 2^4.9 = 269
3600 * 24 * 3 = 2.59M

269/2.59 = 103.
103 reads per second in the last month of fifth year

(t-1)
125 * 2^4 - 125 * 2^3.9 = 134.
134/2.59 = 51.7 reads per second - still very low??
can handle 10KB reads even very easily


--copy dump

# System Design 2 - Q&A site

## Details

### Initial Requirements after 1 year
- 5 Million Questions
- 20 Million answers
- 10 Million users
- View Questions by Recent and Most popular
- Allow users to vote on Questions (and Answers?)
- (Optional - allow users to edit/delete)
- (Optional - do it with hardware bought from a store)

### Data 
#
questions
answers
users
votes

Questions and Answer 2KB on average - can be much more

#### Users
| userid | email | password |
| ------ | ------ | ------ |
| 64-bit integer (8 bytes) | variable length string (250 bytes) | hashed var string (32 bytes) |

#### Questions
| questionid | userid | votes | content |
| ------ | ------ | ------ | ------ |
| 64-bit integer (8 bytes) | 64-bit integer (8 bytes) | 64-bit integer (8 bytes) | variable length string (2KB on avg) |

Question?? should store answers on Question? nah. need it's own votes. nearby would be good though. or at least index the answers by question id and answerid

##### option 1:
#
| shortened-link | original link |
| ------ | ------ |
| variable length string (assume 50 bytes on avg) |  variable length string (assume 2KB on avg) |
##### option2:
#
| shortened-link | original link id  |
| ------ | ------ |
| variable length string (assume 50 bytes on avg) | 64 bit integer (8 bytes) |

| original link id  | original link |
| ------ | ------ |
|  64 bit integer (8 bytes) | variable length string (assume 2KB on avg) |

#### Answers
| answerid | questionid | userid | votes | content |
| ------ | ------ | ------ | ------ | ------ |
| 64-bit integer (8 bytes) | 64-bit integer (8 bytes) | 64-bit integer (8 bytes) | 64-bit integer (8 bytes) | variable length string (2KB on avg) |

#### Question Votes (option to change upord down delete to byte flags 0 neutral, 1 up, -1 down)
| voteid | userid | questionId | upOrDown | deleted |
| ------ | ------ | ------ | ------ | ------ | 
| 64-bit integer (8 bytes) | 64-bit integer (8 bytes) | 64-bit integer (8 bytes) |  bool (1 bit) | bool (1 bit) |

#### Answer Votes 
| voteid | userid | answerId | upOrDown | deleted |
| ------ | ------ | ------ | ------ |------ |
| 64-bit integer (8 bytes) | 64-bit integer (8 bytes) |  variable length string (assume 2KB on avg) | bool (1 bit) | bool (1 bit) |

| shortened-link | metadata/request header/body |
| ------ | ------ |
| 64bit ( 8bytes) | variable length 500-1000 bytes (round up to 2k) |

### Analysis
We are generating about 10M links per day. On a per second basis, that's about 116 requests per second. Even if we say we might much more than this during periods of high load and traffic, it seems to safe to assume that we won't often seek spikes of over 200 requests per second, and even more true for higher rates (300, 400, 500 etc.).

Math for link generations
- serving 10M per day
- 360 s/hour * 24 hours in  day = 8.64 * 10^4 seconds / day
- 10 * 10 ^ 6 requests per day / 8.64 * 10^4 seconds per day = 11.6 * 10 requests per second on avg
We can maybe round it up to 15.0 * 10 (150) or even 500 to handle spikyness/spurtiness

We expect about 100 hits on each generated link on average. And after 2-3 years after initial generation, we won't see anymore requests for a shortened link we created. To start, we'll assume we need to serve anyone of the links we've made over the past 2-3 years. We'll start with 3 years as a high end estimate. 10 million links per day for 3 years come out to about 11 Billion unique links which might need to be serviced. 

Depending on the distribution of when links are requested, it seems there will be some bottle neck on how many links a machine can search for and service. As an exercise, we can asssume that we can put unlimited memory or disk on a single machine. If we keep the links only on one disk and assume a read for a link only takes one IO and a disk can perform 200 IO operations per second, we can serve only about 200 link requests per second. That's at a similar magnitude as link generation. 

Since we expect 100x more reads and write, that seems insufficient. We can add more disks and cores to a machine to get more access, and batch the reads so we get more hits. So if we were somehow able to get 10 links for each read and use 16 disks with 16 cores, we could raise our throughput to 32,000 reads per second for some arbitrary amount of Disk Space. 

In comparison, if we store the links in RAM, main memory references only take 100ns. If we assume one reference for a link lookup in main memory, we can perform 10M references in a second. If we add 16 cores, we can do around 160M lookups. This base magnitude increase in throughput gives us an indication that we should serve as many requests as possible from main memory. On the scale of one machine, RAM lookups outperform disk lookups by far.

So it seems we wish to keep as many of the links in memory as possible. Assuming we have 3 years worth of links (11 billion), with about 2KB data on average, we want to hold about 22TB of data in memory. If we blow that up to 40TB to account for other overhead and estimation errors, we would need about 640 machines to hold this data in memory. And we can easily scale this up and down. 

Math for link requests
- 3 * 365 = 1095 days
- 1095 * 10M = 10,950M = 11.0 B links
- disk ops / sec = 100-200
if batching 10-1 then 1000 disk seeks, 2000 IOPS
if using 16 core computer with 16 disk 32000 IOPS?
- main memory reference = 100 ns 
- 1 / 100 ns = .01 * 10^9 = 1.0 * 10^7 = 10M references per second
- 64GB / 2KB = (2 ^ 20) * (2 ^ 6) KB / 2KB = 2^26KB / 2KB = 2^25KB ~ 32M link records
- 2^36 = 64GB
40 * 2^40 = 40TB
40TB / 64GB = 40 * 2^40 / 2^36 = 40 * 2^4 = 16 * 40 = 640 machines

If we we remove the oldest links, given that we have enough room to store 2-3 years worth of links and we don't expect to see requests for older links, we should be able to keep all 3 years worth of links in memory. 

We can also archive keep all the links on disk as well. And we should in order to make sure they are persisted in case the system goes down. Since we expect to partition the data across 64GB machines, we can expect that each machine can also write the partition's data to disk. Holding 3 years worth of data, if we use a 4TB drive, we should be able to hold 64x more data on hard drive. This should be enough for almost 200 years worth of links. More than enough time to consider changes in the future. We also should be ok servicing any misses in memory from disk since we don't expect links to be hit after 3 years. 

A simple LRU Clock algorithm at a large time scale can be used for retiring the oldest data. With a little configuration I think it's reasonable to expect each node to have enough capacity to serve all read and write requests. We should also add some logging on this system which tracks misses and how much memory is being used. This way we can potentially analyze the data or use it to re-partition the data on the fly.

As far as writes go, we are only writing 10M records per day, less than 150 records per second. With 640 machines, if we assume relatively uniform distribution, we should only need to write to each machine once approximately every 4 seconds. Plenty of capacity given our previously stated disk specs. 

We may also consider how we store the data. We can store the shortened link and the original link together or separately. We may find that many links we generate target the same original link. Given that we can control the generated links to be very small, but the original link is likely to be larger, and possibly much larger, we might save a lot of space by keeping them separate. This would could potentially cut down the number of machines we need, depening on how many times large links are targeted by the same short link. But it may add a decent amount of complexity and latency for the caching mechanism. We think its safer and faster to implement the solution which simply stores the links together, regardless of repeats. If we gain insight in the future that this design would be helpful, we can use it. But it seems a waste to do extra work which may not help at the beginning.

In order to store data about the our link hits, we will keep a log. It should suffice to store the link id and the metadata in the log, which is about 1KB on the high end. In this case, we are assuming we only store successful lookups rather than random links which are not in our system, but maybe some malicious actor tried to use.

In the case, that we want to store all request, invalid or invalid, we can accomplish it with just slightly more space since short links shouldn't be too long either. 

In both cases, we can add time as a 64 bit integer and also simply append it to a disk for storage. If we are ok with sometimes losing data. Because we need to write so many requests to disk, we will be heavily constrained by our disk throughput. So we will need many machines to service very heavy loads of requests. In this case, we don't care too much about latency and reads. We will store it on disk for a while, and we can move read and structure the data at our convenience. We can also accept losing data, so we don't necessarily need a hot standby. This process takes places in isolation of the request task. 




### Components
1. Load Balancer / Server
2. Cache
3. Disk Storage
4. Hit Log

insert diagram here


### Replication and Failure

In order to handle failure, we will use replication. We wilk replicate data within a cluster in a data center. We can replicate across nodes within a datacenter using a Dynamo style.  And we will also replicate clusters aross failure isolated data centers and geographic zones. Replication across geographic regions also has the added benefit of reducing latency on reads. We will use a single leader approach however, since we should be handle writes on a single cluster, and we place less importance on the availability of the link generation component. We can switch over to hot standby if our link leader goes down. And if both goes down, we can absorb this failure in an error budget. If this is not sufficient than we may consider using Paxos to determine new leaders and add in recovery methods for resolving conflicts when fallen nodes come back up. 

Furthermore, a link record is about 2KB on average. So on a machine with 64GB RAM can fit around 32 million link records. So for any amount of links we can fit on RAM, we should have enough capacity to lookup any serve

link requests 
about 100 for every link
over 2-3 years after link generation
unlikely but all links  of past 2-3 could be requested multiple times in a day
should have all of them cached
3 * 365 = 1095 days
1095 * 10M = 10,950M = 11.0 B links

disk ops / sec = 100-200
if batching 10-1 then 1000 disk seeks, 2000 IOPS ?
if using 16 core computer with 16 disk 16000 disk seeks?, 32000 IOPS?
main memory reference = 100 ns ->1 / 100 ns = .01 * 10^9 = 1.0 * 10^7 = 10M references per sec
64GB / 2KB = (2 ^ 20) * (2 ^ 6) KB / 2KB = 2^26KB / 2KB = 2^25KB ~ 32M
11.0 B * 2KB (very high amount of data for link)
about 22. 0 * 10^12bytes ~22TB of data. We can round up to maybe 100TB again.
disk reads 1MB in 30ms
memory reads 1MB in 250 usec

memory reads more than 100 times faster, which means reads 100x more data in same amount of time

so any disk configuration would read data much slower than RAM configuration
And we would need maybe 100x more machines to service all requests from disk.

From this conclusion, it seems safe to assume we need to serve the vast majority of our requests from RAM. We will use disk for persistent storage of all the unique links generated, but it seems we better have any link requested already in RAM when it reaches our url database for lookup.


Using the estimation that links are only serviced for about 2-3 years after generation, we can allocate enough machines which can hold that many links in memory, with some buffer for links which break this rule and other data structures and computation we may need.

So at the high end 3 years * 365 days per year * 10M links per day = 11.0 Billion links.
11.0 B * 4KB (very high amount of data for link)
about 44. 0 * 10^12bytes ~44TB of data. We can round up to maybe 100TB again.

In any case, we need 1,563 64GB machines to cache all this data in main memory. (probably on the high end, but we can scale up or down given production data and performance)

We can propose two options for storing the shortened to original data. One option is to simply store the short links with the original links, which should give us the data shape written above. Option two is to split the storage of the short link and the original link. If we have a large enough proportion of short links pointing to the same original link, we can potentially save a lot of space and therefore machines. This optimization would be slightly more complex than the first version, but cut costs by a lot if the data allows. 

We will assume that the first option is the best way to start rather than making assumptions about how actual links look like in production. Rather than implement a solution we don't know may be useful, it's probably better to solve for the general case and iterate in future.

64 GB = 2^6 GB 
4TB = 2^12 GB
4TB / 64GB ~ 64
since we are partitioning about 3 years worth of data on 64GB machine memory, it should be ok to store the same partitions data on disk 64x. so we given same partitioning, we should be comfortable storing 150 years of data on a 4TB disk.
This should be enough room for all the data we want to store of a partition, including indexes. 

If data is partitioned in a way that the data access is about uniform, then link generation at even 500 links per second (50x higher than avg) should only need to check if a link exists on a node once every 3 seconds. 

So if we assume 200 IOPS for the disk and about 1 or even 10 io to read. We have plenty of capacity to service checks for unique links. We might add a little more overhead for holding trying to commit and abort if another node tries to create the same link. Assuming single leader, shouldn't have too much issue.

Load balancer and request and generate servers can scale up and down to account for more or less load. should need minimal coordination. Assume we can rely on underling data storage on link storage nodes to handle conflicts on links. Abort and return and retry on existing link for newly generated link.

link generation subsystem. how do create fairly unique links? space is very large. shouldn't have issues randomly generating links for even 10 characters. 60^10 = 6.0466176e+17
almost 1 billion billion links. so the space is quite large, if using an appropriately randomized algorithm, very unlikely to collide with another even at 10 characters. can easily add more characters to lower probability.

What's a good way to generate the randomized links though?

storage system - dynamo style storage? will scale up and down automatically, reduces complexity. good for read heavy, write minimal system. Will be great if the partioning works over thousands machines and can easily split when directed, or when issues arise. any key value should only get one write so there isn't much issue with conflicting writes. can probably account for that with error budget and resolve in rare instance with custom code that notifies user and stops serving those links. 
ring partition with consistent hashing?
duplication on "adjacent nodes"
no duplication - rely on replication of system at multiple data center

We only need key-value lookups with some replication. It feels like dynamo style database would work well for the system on a single data center. We only write once (not mutating links), so there shouldn't be much consideration for write conflicts. And Dynamo should handle the replication and failover for us fairly well. If not, we can consider other solutions or more customization on partitioning/re-partitioning and failover.

We can handle high loads of requests by adding a load-balancing system in the front. We can scale as many machines we need to handle large volumes of requests and multiplex them out to the apprpriate storage node.

Caching - we can clean up data using some variation of LRU / CLock which reclaims memory on a regular time scale. 

Replication across geographic regions - In order to bring latency down and also to protect against data center failure - (either connection, media, user input, or other), it seems appropriate to replicate our system across multiple data centers in different geographic regions. 
Latency should see a great improvement for vast majority of reads. Since it's so read-heavy this seems ideal. 
It also seems we can get away with having a single leader system. So the one data center will have the leader, and maybe a relatively close data center will have the hot standby.
Then we can add asyncrounous replications as appropriate.
The hot standby will slow down the writes slightly, but we should have plenty of capacity for it. It may take up to say 25ms for the write request to make a round trip to the hot standby, but this latency should be fairly negligible to the creators of the links, and we can afford much more latency in the link generation process than the hit request process.
Even though this might slow latency, it shouldn't necessarily reduce throughput, as operations can be serviced in isolation for the most part. And all will experience this same latency increase.

If we have a leader failure, the standby can become the new leader. And in the case of both standby and leader failing, we will absorb in the error budget. We could potentially implement a solution which coordinates a new leader and hot standby, but it doesn't seem necessary if we assume that both only go down in a catasrophe. We can also continue to service reads from remaining nodes, and can reach out to users who generate links to do damage control.


As for metdata on the link requests, we can add them to some data store. We can keep a Log of all the requests with meta data. If we can write sequentially fast enough, we can read and process that data a larger time scale at a later data. Else, we can split the machines up to account for all of those. We can easily batch and sequentially write, and if we have need, we can add more disks and machines to handle an increase in the load of writes. 
If desired, we can even drop some requests if it's not critical to save each and every request we've received. 










memory lookups per sec = 


# Dillinger
## _The Last Markdown Editor, Ever_

[![N|Solid](https://cldup.com/dTxpPi9lDf.thumb.png)](https://nodesource.com/products/nsolid)

[![Build Status](https://travis-ci.org/joemccann/dillinger.svg?branch=master)](https://travis-ci.org/joemccann/dillinger)

Dillinger is a cloud-enabled, mobile-ready, offline-storage compatible,
AngularJS-powered HTML5 Markdown editor.

- Type some Markdown on the left
- See HTML in the right
- ✨Magic ✨

## Features

- Import a HTML file and watch it magically convert to Markdown
- Drag and drop images (requires your Dropbox account be linked)
- Import and save files from GitHub, Dropbox, Google Drive and One Drive
- Drag and drop markdown and HTML files into Dillinger
- Export documents as Markdown, HTML and PDF

Markdown is a lightweight markup language based on the formatting conventions
that people naturally use in email.
As [John Gruber] writes on the [Markdown site][df1]

> The overriding design goal for Markdown's
> formatting syntax is to make it as readable
> as possible. The idea is that a
> Markdown-formatted document should be
> publishable as-is, as plain text, without
> looking like it's been marked up with tags
> or formatting instructions.

This text you see here is *actually- written in Markdown! To get a feel
for Markdown's syntax, type some text into the left window and
watch the results in the right.

## Tech

Dillinger uses a number of open source projects to work properly:

- [AngularJS] - HTML enhanced for web apps!
- [Ace Editor] - awesome web-based text editor
- [markdown-it] - Markdown parser done right. Fast and easy to extend.
- [Twitter Bootstrap] - great UI boilerplate for modern web apps
- [node.js] - evented I/O for the backend
- [Express] - fast node.js network app framework [@tjholowaychuk]
- [Gulp] - the streaming build system
- [Breakdance](https://breakdance.github.io/breakdance/) - HTML
to Markdown converter
- [jQuery] - duh

And of course Dillinger itself is open source with a [public repository][dill]
 on GitHub.

## Installation

Dillinger requires [Node.js](https://nodejs.org/) v10+ to run.

Install the dependencies and devDependencies and start the server.

```sh
cd dillinger
npm i
node app
```

For production environments...

```sh
npm install --production
NODE_ENV=production node app
```

## Plugins

Dillinger is currently extended with the following plugins.
Instructions on how to use them in your own application are linked below.

| Plugin | README |
| ------ | ------ |
| Dropbox | [plugins/dropbox/README.md][PlDb] |
| GitHub | [plugins/github/README.md][PlGh] |
| Google Drive | [plugins/googledrive/README.md][PlGd] |
| OneDrive | [plugins/onedrive/README.md][PlOd] |
| Medium | [plugins/medium/README.md][PlMe] |
| Google Analytics | [plugins/googleanalytics/README.md][PlGa] |

## Development

Want to contribute? Great!

Dillinger uses Gulp + Webpack for fast developing.
Make a change in your file and instantaneously see your updates!

Open your favorite Terminal and run these commands.

First Tab:

```sh
node app
```

Second Tab:

```sh
gulp watch
```

(optional) Third:

```sh
karma test
```

#### Building for source

For production release:

```sh
gulp build --prod
```

Generating pre-built zip archives for distribution:

```sh
gulp build dist --prod
```

## Docker

Dillinger is very easy to install and deploy in a Docker container.

By default, the Docker will expose port 8080, so change this within the
Dockerfile if necessary. When ready, simply use the Dockerfile to
build the image.

```sh
cd dillinger
docker build -t <youruser>/dillinger:${package.json.version} .
```

This will create the dillinger image and pull in the necessary dependencies.
Be sure to swap out `${package.json.version}` with the actual
version of Dillinger.

Once done, run the Docker image and map the port to whatever you wish on
your host. In this example, we simply map port 8000 of the host to
port 8080 of the Docker (or whatever port was exposed in the Dockerfile):

```sh
docker run -d -p 8000:8080 --restart=always --cap-add=SYS_ADMIN --name=dillinger <youruser>/dillinger:${package.json.version}
```

> Note: `--capt-add=SYS-ADMIN` is required for PDF rendering.

Verify the deployment by navigating to your server address in
your preferred browser.

```sh
127.0.0.1:8000
```

## License

MIT

**Free Software, Hell Yeah!**

[//]: # (These are reference links used in the body of this note and get stripped out when the markdown processor does its job. There is no need to format nicely because it shouldn't be seen. Thanks SO - http://stackoverflow.com/questions/4823468/store-comments-in-markdown-syntax)

   [dill]: <https://github.com/joemccann/dillinger>
   [git-repo-url]: <https://github.com/joemccann/dillinger.git>
   [john gruber]: <http://daringfireball.net>
   [df1]: <http://daringfireball.net/projects/markdown/>
   [markdown-it]: <https://github.com/markdown-it/markdown-it>
   [Ace Editor]: <http://ace.ajax.org>
   [node.js]: <http://nodejs.org>
   [Twitter Bootstrap]: <http://twitter.github.com/bootstrap/>
   [jQuery]: <http://jquery.com>
   [@tjholowaychuk]: <http://twitter.com/tjholowaychuk>
   [express]: <http://expressjs.com>
   [AngularJS]: <http://angularjs.org>
   [Gulp]: <http://gulpjs.com>

   [PlDb]: <https://github.com/joemccann/dillinger/tree/master/plugins/dropbox/README.md>
   [PlGh]: <https://github.com/joemccann/dillinger/tree/master/plugins/github/README.md>
   [PlGd]: <https://github.com/joemccann/dillinger/tree/master/plugins/googledrive/README.md>
   [PlOd]: <https://github.com/joemccann/dillinger/tree/master/plugins/onedrive/README.md>
   [PlMe]: <https://github.com/joemccann/dillinger/tree/master/plugins/medium/README.md>
   [PlGa]: <https://github.com/RahulHP/dillinger/blob/master/plugins/googleanalytics/README.md>

