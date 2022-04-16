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

### Analysis

We'll first look at the storage requirements for each entity.

Questions: 5 Million, about 2KB avg. 5 * 10^6 * 2 * 10^3 bytes = 10 * 10 ^ 9 bytes ~ 10GB
Answers: 20 Million, about 2KB avg. 20 * 10^6 * 2 * 10^3 bytes = 40 * 10 ^ 9 bytes ~ 40 GB
Users: 10 Million, about 300 bytes max. 10 * 10^6 *  300 bytes = 3000 * 10 ^ 6 bytes ~ 3GB
Question Votes: If we say one for each read of a question, we get 500 Million in a year. Which is about 50 votes per user. Totally possible, thought it seems unlikely. 500 * 10^6 * 32 bytes = 1600 * 10^6 bytes ~ 1.6 GB
Answer Votes: Assuming one vote per question, same analysis as Question votes. ~ 1.6 GB
Total storage: 10GB + 40GB + 3 GB + 1.6GB + 1.6GB ~ 57GB
This is low enough it could even fit on one machine in RAM.

Since disk is cheaper and has lower throughput, let's consider serving all our data from disk first. We can store all the data on one disk. It would be less than 1TB, which is a relatively common hardware component. However, the throughput for one disk would only be about 200 IOPS max. So how many operations could we do, and how much do we need?

~200 IOPS for one disk, 100 on the low end
~Lets assume reading and writing per disk are roughly equivalent for same data. 
~ 1MB in 30ms. 1000 / 30 = 33.3 MB per second.

We expect 10 Million users, 5 Million questions, 20 Million Answers, and  500 Million reads of questions. 

User data is about 300 bytes max. We might save even closer to 50 or 100 on average. 100 bytes * 10 Million ~ 1GB. We use 3GB for a high estimate. We expect 10 Million user / 3.15 Million seconds = 3.17 new users per second. So we should expect to write a 4-5 new users' data every second.

We expect 5 Million Questions at about 2KB per question. 5 Million questions / 3.15 Million second = 1.59 new questions per second.

By same math we expect 6.35 new answers per second and 159. reads of questions per second.

So about 13 new writes of 2KB each per second. And 160 reads of 2KB each. If we aggressively round up, we might say we could spike to 5x that and get around 65 writes and 800 reads. 

Since we have many reads, it seems like we would orgazine the data in some type of b-tree, which would keep our lookups to a relatively low number as it grows logarithmically. 

50^4 = 6.25M
50^5 = 312M
50^6 = 15.6B

With only a few levels in our b-tree of a fanout of fifty, we can expect 4-5 disk accesses for a table with millions of records. If we maintain indices on the table, than we might do another 4-5 for every write, but might not add so much on reads.
That's another 5x on disk acces we expect. Where before we had abou 900 disk accesses on peak traffic, we might expect 4500 disk accesses per second on our b-tree structured data.

We might be able to get away with this if we parallelize all our disk operations on 16 or even 32 core machines with as many disks, and assume we can push 200 accesses per second per disk. 

Luckily, we expect most of our reads to already be in the buffer pages on RAM, if we allocate enough. Even if we only allocate enough memory for 50% of all we plan to store, we should expect to get around 90% hit rate on data in memory. [a bit hand wavy, but assume that most popular and most recent questions dominate the reads by a lot. we can adjust if not the case], and can use expect to buffer our writes in memory as well, until pages are written to disk, with only sequential WAL writes needing to be written immediately. Thus it seems reasonable to assume that we can bring our disk accesses to around the avg 200 per second range, and at least under 1000, which should be manageable given multiple core and disks on a single machine. 

So given one machine with a possible a few cores, 64GB RAM, and 4 TB hard drive. We should have more than enough throughput and capacity to manage simple writing and disk accesses on questions and answers.

However in this case, we also are trying to give the users to view the Questions by most recent and most popular. And we need to implement a voting system to organize the list by most popular. So we need to keep track of how many votes a question(and maybe answer) receive. And we want users to only have one vote per question/answer.

If this is the case, we have to keep track of all votes. We don't really know how many votes there will be. All 10 Million users can vote for each question. And we probably expect them to vote for only one answer per question. However, we also know there's only 500 Million reads. Assuming users generally read a question before voting for it (maybe they only read from a feed sometimes), we can estimate the votes as about the same as the 500M reads of questions. This may be too high if users don't often vote or often visit the same page which they can only vote once, but serves as a starting point.

If there are such votes we can use the previous math to estimate that there will be about 169 votes for questions and 169 votes for answers every second. This is pretty heavy, compared to load we calculate previous. If we consider that we need to index the votes by question and answer and user, we add additional overhead. And we also have to count the number of votes a question (or answer) has in order to rank it for the most popular view.

We want to support user voting and ranking of items. This means we need to store a vote on an item by user, and we need to be able to quickly find if a user's vote and also count all the votes for an item. 

In terms of storage we could consider a log structured storage to speed up writes, but we could hurt the read performance a lot. And it would be even worse if users are not really voting as often as they read questions. So a b-tree seems still be a good approach, and we can continue assuming we use postgresql or a similar storage engine. So we will want to store the votes by their question (and answer if voting for answer) and user id. However, in order to determine the popularity of an item, we would still have to count all the votes. If votes are small and votes are stored together, then this counting process would be fairly quick. On the other hand, if we keep an index of the votes and use transactions to increment the count separately, we might get an improvement on counting votes, though have to store more.

Option 1: Store in b-tree. key is  questionid-userid
advantage: don't have to keep separate index.
disadvantage: doesn't seem to be any. It only ever needs to be found in this way. unless we kept a log of the votes. And update the index only every so often?

Option 1a: Don't store count.
advantage: no extra storage.
Disadvantage: must read all vote records and do count calculation every time there's a ranking query. must also count all vote records for an item when there's a read on the site to let users know how many votes an item has. 

Option 1b: Store count on the questions/answer record.
advantage: have the count colocated with the question/answer
disadvantage: need to re-write the whole question/answer, every time the count is updated,
need to update the indices fairly often as answers are updated to keep track of the location. need to keep another index of items in order of most votes

Option 2: Store count separately from record. orderd by question / answer id
advantage: only need to re-write a small amount of data on each vote. Can separate this lower priority item from the main retrieval of data. User's don't necessarily need to have most accurate count all the time.
disadvantage: extra latency to retrieve the count from a different machine.

Option 2a:  keep an index ordered by count.
advantage: easy to reload the order by count as its save in disk. can use postgresql buffer pages as natural caching of items in memory.
disadvantage. Many re-writes in memory as counts are updated. Need to track as the original count get's updated often.


Option 2b: keep the order by count as an in memory structure. like a skip list or a balanced tree. (only 5Million * 100 bytes depending on type of structure ~ 500MB of RAM). 200 bytes if including some truncated portion of the title. (1GB of RAM). (Not necessary for answers but maybe about 2GB of RAM.)
advantage: fast. Can quickly update and read the count data.
disadvantage: take a little to build on starts/restarts. 
alternative: keep the in memory structure while periodically dumping to disk? or keep in memory structure and rebuild from stored data? only a few extra GB RAM. but potentially slows the index process as it writes. Well technically don't have to read it back. As long as commit in log happens, we can assume it's durable and just write to memory.

Latency within same datacenter is about 500usec. And being ables to read and write 1 MB over network in 10ms. That is 100MB per sec. Given each request is less than 100 bytes. We can over a million requests. If responses are about 10Kb max for a whole page of data. then we can do 10K responses per second. This is much higher than the load we calculated previously with only hundreds of qps on avg and maybe a thousand as an unlikely high end. 

If we replicate the data to another node, we can expect to add < 100ms latency in the same country/geographical region. So overall the latency should not exceed 250ms ever. 

So my favored design is to keep one machine which stores the questions and answers and users. (users could be split off if its too cumbersome.). The other machine stores the votes, and keeps a version of the questions ordered by votes. [we can also add a count feature to a b-tree in memory. This would allow us to easily page and range on items]

[node has [low, high, count] then we can easily determine how to get page 1,2,3,...etc with 50-100 etc.]

if we were to scale this continuously, we would store diff range of counts on diff nodes, and we could pass the data between each other. But this wouldn't be necessary until a few decades worth of data had been reached.

Final component? besides voting and ordering? should we store the most recent as well?
It doesn't really change much. Given that it doesn't change and only needs a few GB data. It would be best to store it on the first node if possible. We could also just page by it's id and re-order on the front-end. Or we could partition the data of the first node as we scale up. Then we could generally partition it as more nodes come up so that data is either on the most recent node or only one hop away. 


One data flow. Two designs that can be added.

User Sign in
Server receives email and password.
Forward to storage node with Users.
Hash password.
Check.
If success: create token for session and return.
else return failure.

Other -
receive token with data. Check session agains existing sessions.
Block if not ok.


Read request on question
Receive question ID and user ID on server. Forward to main storage node and votes node. 
-question storage.
Read the question data from the db. hopefully in the buffer page. else goes to disk.
-answer storage
Grab all answers related to the question. would be nice to store it next to. but if using postgresql maybe not possible.
-Vote storage
read vote count for question (and maybe answers)
read if and how user voted for these.

original node receives all these and forwards to caller/front end.

Read request for most popular - better if most popular only has ten or so pages
-server receives this request. which is the page of most popular desired. 
forwards to votes node. Votes node traverses in memory structure and grabs page range of items (can be any base 25,50,100 etc). 
Returns all array of page data with ranks, and votes



Read request for most recent.
Server receives request.
Forwards to machine with posted date paging struture. 
Machine traverses memory structure and returns to server and user.

Post new question
Main server receives data with user, and questiond data.
Forwards to question storage server.
Data is stored. Indexes are updated for id and most recent.
Add count to vote server with 0 count.
Replicate synchronously to another isolated data center nearby.
Replicate asyncronously to other data centers / zone.
Return success to user.

Post new answer
Main server receives data with user, question id, and answer data.
Forwards to main storage node or whichever has questions.
Data is stored.
Add to vote node with 0 count.
Replicate synchronously to another isolated data center nearby.
Replicate asyncronously to other data centers / zone.
Return success.

Vote on question.
Forward to storage node for votes.
Write vote to disk. (Transaction updates vote and count)
Update counter in memory on success.
Replicate synchronously to another isolated data center nearby.
Replicate asyncronously to other data centers / zone.
Return success.


Vote on answer
Forward to storage node for votes.
Write vote to disk. (Transaction updates vote and count)
Update counter in memory on success.
Replicate synchronously to another isolated data center nearby.
Replicate asyncronously to other data centers / zone.
Return success.





Design one. Have data on a few machines. Use postgresql btrees, serialized??

500M / 365 = 1.3698630137 M questions read per day
8.64 * 10^3 sec per day
8.64 * 10^3 sec per day * 365 day = 3.15 * 10^6
500 * 10^6 / 3.15 * 10^6 = 1.59 * 10^2 ~ 159 questions read per second (queries per second) qps
159 qps on questions

with a btree of fanout 50 -> 50^4 = 6.25 Million
easy to fit 5 million questions on a tree with around 4 levels
anyway 2KB questions 5M ~ 10GB. would easily fit in ram. double can double it even to fit the entire btree and only 20GB. shouldn't need that much though

how to store by recent and by top vote??
seperate index??
two indices - one for votes and one for recent
how often does the index get updated?

Design one:
postrgresql instance with #votes on questions and answers. 
Questions has indices on recent/time posted and number of votes.
advantage: simple to implement
disadvantage: many more write operations and maintenance for indices

Design two: separate component for votes and tracking most popular questions (can also put a limit on most popular questions)

Design three:


Option one: 
postgresql instance with questions, answers, users, votes
[advantage - simple]
[disadvantage - may not be able to keep track of votes?, many writes through postgresql, joins?]
[have to run a new query to keep track of updates to top pages]

Question: how many votes will we have to write???

Option two:
postgresql instance with separate store for votes and which also holds a RAM version of most popular questions with b-tree?
why? because having to update a question everytime it's voted on seems overkill.
It would be better to just store the vote for a user on a question and join it somewhere else. Also want to avoid joining all the time, when each user only needs a small amount of data on each join. (take a while top populate??)
so keep a component which stores the votes?? but wait don't i need votes to show on most recent?? not necessarily? I can probably get a away with not showing it on most recent
.. but if i want to show on most recent, I guess I have to colocate??
no I can afford a network hop most likely. Latency will be worse. but it should be acceptable. only adds 500usec. Can store almost all of it in memory??

500M * 32 bytes = 16GB. if storing answers, just about doubles. 24GB total. pretty doable it seems. can store less if i'm just doing questions and vote count. With other stuff cached.

need to write 1B votes. 500M for questions, 500M for answers. probably less if you save one 1/10 reads have a vote attached. 1/10 have 2 votes attached? only 100M













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

