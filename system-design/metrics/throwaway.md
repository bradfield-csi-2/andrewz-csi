
In the future, we may find that because of increase in scale and/or after storing a few years data, we are having trouble storing all the data. In this case, we really need to be stingy with our storage and use as little as possible. And so we can make use of our

Do we need SSD's, No. We can use log structured merge tree. It should be able to handle 50KB per second of writes. Though we may have very long compaction periods that back us up. We can write 1MB/30ms so more than 30MB every second. At larger compaction levels, we might be writing quite a lot of data so it would be nice to have SSD's but it should be okay to compact for a few seconds and just append to the log in parallel. Or if we use multiple disks, we can keep appending adding new data in parallel, and run compaction rounds at opportune times. 


Either way, it would be rather cheap in general. 


3TB on disk seems fairly reasonable. We might double this in order to do some more interesting processing. 6-7TB seems reasonable for disk storage.

However if this blows up that's 600-700TB for 100x the data points. We might have consider storing less data on disk in that case.



Well we would consider that we could get good compression of our storage. Given we are storing columnwise data and have a good sense of how the data looks like, we should be able to compress it to 50% or even 25% of it's orginal size. 

Hopefully with that much scale up in our system, we would have a decent budget for that kind of disk storage.

Other that, we might consider mostly cold storage for our data. 
For processing we could keep most everything in cold storage. And we can keep the time series data on disk? or we could keep a months worth of data on disk and everything else in cold storage?? Would be nice if we could retrieve the data within miliseconds. We can probably pay a few thousand dollars to get enough capacity for a year even at 100x. 200-250TB storage. At 50% compression that's only 100-125TB

The rest we can process in cold storage. We may need a little bit of extra disk space to help us process the data comning from cold storage but should be find with one extra large hard drive.



Anyway. If it grows beyond 100x from the original estimate, we may need to consider throwing away the data. We can maybe keep some in cold storage, but it might be better to just keep some summary stats and keep a small sample of the data. At 100x, it might be ok to keep 5-10 years. Older than that is probably less useful.

Storage
A log structured, column store seems to be the most useful for storing time series data. 
1. The data comes in mostly sorted already.
2. We are mostly writing.
3. Recent data is read more often is and is easier to find.
4. We only analyze one or two columns at a time.

Batching scheme. 
To support the aggregate functions we need, we only need a minimal amount of space for sum, count, average, min, max.
However, median and percentiles require us to keep the entire data set sorted for accurate calculations.

But rather than do all the processing at the end in one batch. We can amortize the cost over time, by sorting the data in batches over regular periods and merging many smaller batches into one large batch.


So every hours data can be sorted. Then all the hours data is merged at the end of the day. At the end of each weak and month, the days data is merged.
Then at the end of a quarter and year, the month or quarter data is merged.

This allows us to only do a full sort on small batches at a time. At the end we are mostly reading and copying data which is inexpensive process.

Further more, as we are sorting and merging, we can add prefix counts to large blocks of data. This allows us to quickly find the data points we want.

30 machines * 30 metrics * 3 apps = 3k metrics
there are 60 * 24 minutes in a day -> 1440 minutes in a day
525K minutes in a year
526k times 100 bytes is 55MB only 
even times 100x is only 5.5GB a year

there are 24 hours in a day. 8760 hours in a year
8760 * 100 bytes is 876MB a year
100x is 87.6GB a year 


365 days in a year
365 * 100bytes = 36.5KB per year
100x is 3.6MB year year

should be plent of room to store summary statistics even by the minute.
But maybe it's best to just get by the hour. If by the minute, almost would rather just reload the second and do the entire analysis.

As far as latency, once the minute is sorted, then we can merge all the other minutes into one hour. The minutes should take almost no time to sort. Maybe 36 microseconds max. With parallel computing and around 9 cross sections, that's less than a millisecond to process. Then merging the minutes to an hour is simply more copying. It takes 30ms to read 1MB from disk, the majority of the time will be taken by thiss.

One hour is 50kb * 3600 = ~180MB. With parellel processing it should only take about 300-500ms.

Mergeing one day's worth 50KB * 86.4k = ~ 50KB * 100k = 5GB.
Copying from disk takes around 10 seconds.






Then merging one week will take over a minute.
Mergong one month will take five minutes.
Merging one year will take abount an hour.

1. Talk about Merging and Batching and Preprocessing
2. Talk about resevoir sampling
3. Talk about compression schemes

Compression Compression Compression

