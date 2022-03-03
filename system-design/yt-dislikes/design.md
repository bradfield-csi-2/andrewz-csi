## Youtube Dislike Esitmation

### Requirements
1. Collect likes and dislikes and estimate total dislikes of page
2. Support 1 Billion users for around 30 Billion videos on youtube

### Summary
- Read-heavy, users often view a video without liking or disliking
- Average like ratio seems around 4% of views anecdotally
- 25% ratio like to views would be super elite
- Since we can't really know the like count with just our user information, we might as well give up on total accuracy from the start
- Store just the count of likes and dislikes from our users on a video
- Try to make it as correct and efficient as possible
- Use ratio of user likes to real likes to estimate true dislikes
- Since we are only keeping a count, we can probably store just a few bytes for each video id.


### Design

Our system's main function is providing users a reasonable estimate of the total dislikes a YouTube video has. Many users desire this function on the Youtube even though it is no longer supported on the app. So we should first answer how we can do this? 

Well since Youtube provides the total likes a video has, we can use our user base as an indicative sample of how many likes and dislikes a video receives in a sample population. Basically we can compare how many of our users have liked a video to the total likes reported by Youtube and extrapolate the total dislikes given our user dislikes.

The key items we will need to address are the volume of reads and writes we are receiving from our 1 Billion user base and maintaining an accurate count for 30 Youtube videos.

Given these, it seems best to avoid keeping data of which videos each user liked in order to significantly minimize the volume of data we need to read/write. And since it is impossible for our estimate to be 100% accurate anyway, it seems reasonable to give up a little more accuracy for system performance.

Therefore, we will keep only a count of likes and dislikes on each video by our users. (We can optionally store the total like count, but it seems reasonable to offload this to the actual youtube data that our clients are receiving. Actually I'm not clear whether we can parse this data using an extension but it seems reasonable and would save us having to do any unnecessary compute with our own hardware). And will also listen to events in the user's browser to determine when they like/dislike a video (I suppose we can't do this for mobile but oh well). 

Since we are only tracking counts, we would like to introduce some level of idempotency into the system to improve the accuracy of the counts we collect. Therfore, we will program our client side extension to buffer writes and coordinate with our backend to have them accepted. This also allows us to throttle/apply backpressure to updates and limit bad actors.

Essentially, our clients will write updates to the client storage and then request that they be accepted by our backend. Once they are committed to the backend, they can stop trying to send the update. Otherwise they should keep retrying. This gives us some relative insurance that all client likes/dislikes will be correctly counted. (see message queue for further details)

This could be included as part of our database, but it feels better to separate this functionality into a message queue, this way we can separate the responsibility of accepting and replicating messages from our database component. 

As for our datastore, we have a very simple data pattern that is essentially a big hashmap. If we are careful with using only atomic increments and decrements and tracking which updates have been flushed to disk, we can concurrenly count the likes and dislikes without much issue. We don't have to worry about undos since there is no reason to abort a transaction (no conflicts, taken care of by message queue coordination), and we just need to be careful with redos by not flushing dirty pages until all transactions before the last to write to a page have completed.

In this case, we would ideally use a memory database like Redis. But we since we have so many videos, it will probably be cheaper to store most of it on disk. Therefore, I am not clear if we can use Redis. A data store with some type of buffer page manager would be best. Though I'm not sure if one might exist on the market, the features are simple enough that we can probably create our own implementation if necessary. 

With this system, we could reasonably serve these functions from one geographic location, even over the world, by accepting higher latency, probablility of dropped or requests/messages, and traffic around our main node. However, it seems we could improve all these by splitting our service into multiple nodes to service each individual region.

With our counts only data structure, we can even use Conflict Free Replicated Datatypes to implement a fairly efficient and consistent leaderless system. We would need to add more space so each partition can track it's counts and also have two numbers of increments and decrements, but this seems a good tradeoff to make. We also should be cognizant of users which hop between regions while coordinating their updates to the message queue, but this should be a relatively rare and simple edge case to handle.


#### Components
1. Message Queue - for extension to load updates to so they can be pulled by server
2. Notication Mailbox - for users to know their messages have been successfully submitted and will be consumed by the system soon
3. Hashmap DB of video and counts - store total user likes/dislikes
4. Authentication/Identity Server - make sure we don't collect data from bad actors


#### Count Data 
| VideoId | Num Count Blocks |  Counts Data| 
| ------ | ------ | ------ |
| varchar (15 bytes max) | 8-bit integer (1 byte) | variable array each element having 4 64-bit integers { #likes , #like undos , #dislikes , #dislike undos} |

Each video has has an array which is a CRDT (conflict free replicated data type.) We collect the number of likes and undo likes and dislikes and undo dislikes from each user. We can easily partition these by region and not have to worry about overwriting other regional data because each region handles it's own count and gossips with the other regions. Then the total is simply the sum of all the counts.

The total size of this might be 16 + 32*5 = 176 bytes for five regions. We will round to 200 bytes.

#### Estimates and Calculations
1 Billion users
30 Billion videos

With 30 Billion videos and 200 bytes data each, that's at least 6 TB of data we need to store.

This could fit on a hard drive of 6TB, or we could split it into 2TB SSD (or less if they have larger sizes). We could also fit it into a few TB of RAM (or one really huge memory machine) on different machines, which is helpful for keeping data in memory.

With 1 Billion users globally we expect a third of them to be asleep at any one time. But even if we heavy concentration of users in on region, let's say we only have 800 Million users awake at one time.

If they view 10 videos in one hour (a lot), that's 8 Billion views in one hour. On a per second second basis that 8 Billion / 3600 = 2.2 Million views per second. 

It seems the average like/dislike is around 4% of views anectdotally, but let's give 25% just to be safe (an because we can). So we need to process .55 Million updates per second potentially. [https://programminginsider.com/which-is-more-important-youtube-views-or-youtube-likes/] [https://pixelvalleystudio.com/pmf-articles/4-key-youtube-channel-statistics-and-how-to-calculate-them ] [ https://pex.com/blog/analysis-of-user-engagement-videos-worlds-biggest-social-sites/ ]

That means up to 550K writes to the queue and 550K writes to the notification mailbox and 550K writes to the count database per second at high load. With 2.2M reads from the Count DB.

If we split this into 5 regions, that's 440K reads and 110K writes on each node spread evenly. If we assume one regions is especially large then maybe one region can be 880K reads and 220K writes per second.

And if we assume that the gossiping is extremely active and each write is propagated to the other nodes, then we might receive the full amount of writes to each node (550K). However, since the writes from other partitions don't need to be written to log or anything, then can essentially count as reads. So we would have on the order of 1M+ reads to each node. 

We will also assume some things about some basic hardware performance.
10ms to read 1 MB using a 100MB NIC
1ms to read 1MB using 1 GB NIC
100 usec for 10 GB NIC - reasonable couple hundred dollars in price
10 usec for 100 GB NIC - expensive $2k ?

//Samsung EVO Plus SSD: ~ $200 https://www.samsung.com/us/computing/memory-storage/solid-state-drives/ssd-970-evo-nvme-m-2-1tb-mz-v7e1t0bw/#specs 
//using 32 cores 
up to 600K IOPS Random read
up to 550 IOPS random write
read up to 3500 MB/s sequentially
write up to 3300 MB/s sequentially
5 years or 1200 TBW for 2TB model

Large CPU with 128 cores and pcie slots - $10K at high end? (AMD EPYC server)
1TB ram $10k - seen on newegg (4x DDR4 256GB kit)

http://highscalability.com/numbers-everyone-should-know


### Count Data Store
![](YT-Redis-Write-2.jpg)

As mentioned in the begininning, the ideal store for this system would be a hashmap on disk, where we write the data to pages/buckets of our drive. We generally just want to read and update the items in place so if we were to keep it entirely in memory, it would be very performant. But we need some type of durability mechanism to ensure we can recover our state from any failure. 

Simply copying the hasmap structure to disk seems reasonable. It is the most memory efficient while giving us very fast access with as few IOs as possible.
Furthermore, we don't necessary need to flush every change to disk right away. If we use an AOF/WAL to append the changes we want, we can update the counts in memory and recover from our last saved state in case of failure. 

Of course, with this volume of data and writes, we should still flush/snapshot our data occassionally. So if we organize our hashmap into buffer pages, we would have a pretty efficient method of blocking data into flushable chunks. And we can customize our fsync policy to flush data to disk at convenient times. It also gives us the flexibility to use really large disks and much less RAM to accomodate different types of hardare configurations if we need. As long as we can tune our buffer page manager, we should be able to get away with paying for a lot less RAM then would be needed to fit the entire data set.

If we use an AOF/WAL we can quickly commit and make changes durable while simply updating the item in memory with atomic increments. Then occasional (once or twice a day) flushing of pages to disk would be enough to keep our truncate and manage our log while also saving our state. (We might also consider continuously re-writing the AOF (merging increments to same video) with occasional snapshots).

If we have had a large enough budget for hardware (around $100k for each region/node), then this would probably be the simplest and most performant setup. However, we can get by with much less, relying on SSD and the access pattern of our data. Most likely, the most popular videos and will be memory if we manage our buffer pages properly. And if not, SSD's should have the capacity to serve reads and writes with tens of microseconds. This could wear out our SSD's much faster, but given that the price of equivalent RAM size is around 100x SSD, we can accept the cost of replacing an SSD drive a few times per year.

We can ask our database server to continually pull new update from the message queue to put into our data store. 

### Message Queue and Mailbox
![](YT-MQ-1.jpg)
![](YT-MQ-2.jpg)
![](YT-MQ-3.jpg)
![](YT-MQ-4.jpg)
![](YT-MQ-5.jpg)
![](YT-MQ-6.jpg)


The purpose of the message queue component is to coordinate updates with the extensions. By assigning space and ids to a user, we are able to control bad actors and to maintain a fairly accurate count of likes/dislikes and actions.

We do this by having the extensions request a message id from the queue when they want to submit an action/update. First, the request is checked by authentication server to identify if the requester is a user on our system. Then it is forwarded to the queue. When the message queue receives the request it assigns a messageId and space in it's buffered transactions to the potential update, and returns it to the requester. It also assigns the requester's user id to the message.

At this point, the queue will buffer the transactions in memory for a short time before committing them to permanent storage where they will become an offical message that needs to be processed by the main db/count system.

When the requester receives the message id, it then sends another request to the queue with the id and the action/update to submit. The auth server verifies the requester again and forwards. Then the queue, tries to match the message id to it's buffered transactions. If the messageId exists and matches the user, the action is entered to the buffered transactions.

When buffered transactions get committed as committed messages, the queue will put status messages for each transaction in the requester's mailbox(can exist on same machine or queue system). If the requester didn't send the update in time, then the transaction is unsuccessful, else it was committed and was successful.

The requester can pull from the mailbox to see if the message it submitted or was planning to submit has been processed succefully or not. This let's the requester know if they should keep retrying to send the update, or if they need to start over and request a new message id from the queue.

If this is done carefully, user's shouldn't have a problem with redundant or lost submissions. Though there is always a chance that they lose their data and submission while it's saved in their machine's memory or they accidently clear their browser data.

If we want to be super careful, we can add some more safeguards so that user's have to keep pulling from the same queue (maybe they move regions mid submission). Or we can replicate mailbox messages which aren't cleared (checked) for a while to other regions. And use some gossip protocol to replicate and coordinate the message availablility between regions.

We also may want to add enough capacity for message queues to keep a day or two's worth of messages in storage. This way, we can have a rotating band of smaller data stores which regularly wake up and consume the messages at different times. This allows us to add replication and recovery, but not have to keep the replicas on all the time.

For the Mailbox we can use Redis as our Data Store Engine as notifications are transient.

### Conflict-Free Replicated Data Type
With our nodes being spread across regions we would typically need a way for them to decide who is the leader and direct all the writes to that node. In this case, we have decided to use a leaderless architecture because the volume of writes is very large, the users are spread out very far geographically, and the character of our data is convenient for each node to handle writes conflict free.

Because we are only doing increments, we can make use of conflict-free replicated data types. Each node will therefore handle incrementing the count (and undo count) within their own region. This gives us the total of the region. Then they gossip among themselves so that each will eventually receive the counts of all the other regions. And the total count is simply the sum of all. 

Because it is increment only, each node knows that it simply has to update a regions count with the largest one it receives. And it is guaranteed to always move forward to the most updated state it has received. And it cannot conflict with other regions since each region only increments its assigned count.

This allows us to accept writes close to a region and reduce the latency it takes to accept and process writes. But it has the downside of short term inconsistencies between regions as they gossip to receive each other's writes. This seems to be a pretty good tradeoff since we would have inconsistincies with a single leader as well, because it would take a while to propagate changes from the leader to the followers, and estimating the true dislike can't be 100% accurate anyway.

Another issue that arises is the way inconsistincies could arise in a failed node which recovers. Because we are not logging the updates to our crdts received in the gossip protocol, some updates might be lost when the node goes down and it wasn't flushed to the drive. When it comes back online, it will have to gossip with the other nodes to receive and check the most up to date data from each region. If we're smart about it, we can keep checking with the local replica for a while before coming fully back online, only accepting data it has gossipped with other nodes about since recovery began.
, but eventually converges

![](YT-CRDT-1.jpg)

