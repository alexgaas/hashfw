## Robin Hood vs Linear Probing hash tables
The purpose of these experiments is to compare the lookup time of the Robin Hood and Linear Probing implementations of 
Open Addressing hash tables. The Robin Hood hash table is designed to have better performance at high load factors 
which is why the comparison focuses on tests with a high load factor.

#### Contents:
1. [Open Addressing](#open-addressing)
2. [Linear Probing](#linear-probing)
3. [Robin Hood Hashing](#robin-hood-hashing)
4. [Load factor](#load-factor)
5. [Hit rate](#hit-rate)
6. [Linear Probing lookup](#linear-probing-lookup)
7. [Robin Hood lookup](#robin-hood-lookup)
8. [Dataset for comparison](#dataset-for-comparison)
9. [Results](#results)

### Open Addressing
Open Addressing is a type of hash table where collisions are always handled by placing the data somewhere else in the 
already allocated array instead of allocating more memory for the colliding data (for example via a linked list). 
The different ways of handling the collisions is the main difference between different types of Open Addressing hash tables.
Open Addressing is used when memory is limited since there is no extra memory usage to know where the colliding data is located. 
The drawback of Open Addressing is that when the hash table is nearly full then inserting and looking up data in the hash table can
slow down as the time to find an empty spot for the data is increasingly hard as collisions become more common.

### Linear Probing
Linear Probing is the simplest form of Open Addressing where collisions are handled by incrementing the index until an empty 
spot is found for the data. This is very simple to implement but has the problem that the data can easily **start "clumping" 
together at one part of the array due to collisions increasing** the chance of further collisions near the same index. 
When the array is nearly full the performance of inserting and looking up data will decrease since more indexes that already 
contain data have to be skipped over leading to the data being placed far from the original hash index.

### Robin Hood Hashing
Robin Hood Hashing is an extension of Linear Probing that attempts to position the data in the hash tables array **more fairly** 
by keeping track of how far from the intended location the data is placed in the array due to collisions. This distance from the intended
location to where the data is placed is called the **"probe sequence length" (PSL)**. 
When collisions occur while inserting data the algorithm starts out like Linear Probing by simply incrementing the index but 
when the data at the current index has a lower PSL than the amount of times the index has been incremented then the data is swapped
with the data at the current index and the algorithm continues finding a place for the data that was replaced. The PSL can 
then be used to reduce the amount of data that has to be searched through if the data that is being searched for does not exist. 
If the PSL of the data at the current index is lower than the length of the current search then the data can not be in the array
as if it was then the data at the current index would have been replaced by the data that is being searched for. 
This means the search can be stopped early to save time.

### Load Factor
Hash tables have a "Load factor" which is an important variable for Robin Hood hash tables and this means how full the array 
currently is on a scale from 0 to 1. An empty array would have a load factor of 0.0, a half-full array would have a load factor of 0.5,
and a full array would have a load factor of 1.0. A technical description can be found in which describes load factor as 
the "ratio of the number of records stored to the total capacity of the memory.".

### Hit rate
Another variable that can affect performance is the "hit rate" which simply means what percentage of lookups are of keys 
that exist in the hash table. For example if half of all lookups are of keys that exist then the hit rate is 50%, and if 
all lookups are of keys that exist then the hit rate is 100%.

### Linear Probing lookup
The lookup algorithm for Linear Probing calculates a hash from the key and then uses the remainder of dividing the hash 
by the length of the array as the index to start the search. It then compares the key of the node at the current index to 
the key that is being searched for and if the key does not match then the search continues on the next index. If the current 
node does not hold any data then null is returned to signal that no data with the given key exists in the array. If the 
key of the node matches what is being searched for then the data is returned.

```text
Input: key ▷ The identifier of the data

1: h ← hash(key)
2: len ← length of the map
3: index ← h%len
4: i ← 0
5: while i < len do
6:      element_index ← index + i
7:      if element_index ≥ len then
8:          element_index ← element_index − len
9:      end if
10:     n ← node at element_index
11:     if n = null then
12:         return null
13:     end if
14:     if the key of n = key then
15:         return data
16:     end if
17:     i ← i + 1
18: end while
19: return null
```

### Robin Hood lookup
The lookup algorithm for Robin Hood is similar to the one in Linear Probing except that it returns if the **PSL** of the 
current node is smaller than the search length.

```text
Input: key ▷ The identifier of the data

1: h ← hash(key)
2: len ← length of the map
3: index ← h%len
4: i ← 0
5: psl ← 0
6: while i < len do
7:      element_index ← index + i
8:      if element_index ≥ len then
9:          element_index ← element_index − len
10:     end if
11:     n ← node at element_index
12:     if n = null or psl of n < psl then
13:         return null
14:     end if
15:     if the key of n = key then
16:         return data
17:     end if
18:     i ← i + 1
19:     psl ← psl + 1
20: end while
21: return null
```

### Dataset for comparison

The benchmark tests use randomly generated string keys with the format `key_{index}_{random_int64}`. Tests are conducted with:
- **10,000 elements** pre-populated for lookup tests
- **Variable sizes** for insert tests (scales with benchmark iterations)
- **~45% load factor** for high load tests (50% threshold triggers resize)

Run benchmarks with:
```bash
go test ./rhlp -bench=. -benchmem
```

### Results

Benchmark results on Apple M4 (arm64):

| Benchmark | Linear Probing | Robin Hood | Notes |
|-----------|----------------|------------|-------|
| Insert | 1718 ns/op | 2284 ns/op | LP faster due to simpler logic |
| Lookup (100% hit) | 4.9 ns/op | 5.2 ns/op | Similar performance |
| Lookup (0% hit - miss) | 19.5 ns/op | 198 ns/op | RH early termination helps at high load |
| High Load Factor | 1089 ns/op | 1253 ns/op | At ~45% load, LP ~13% faster |

**Key Observations:**

1. **Insert Performance**: Linear Probing is ~25% faster on inserts because Robin Hood requires additional swapping operations when "stealing" from rich entries.

2. **Lookup Hits**: Both implementations perform similarly (~5 ns/op) when the key exists, as both find the key quickly.

3. **Lookup Misses**: The Robin Hood early termination based on PSL can provide significant benefits at high load factors by avoiding full table scans.

4. **PSL Statistics** (at 24% load factor with 1000 elements):
   - Max PSL: 2
   - Avg PSL: 0.15

   This shows Robin Hood maintains very low variance in probe lengths.

**When to use each:**

- **Linear Probing**: Prefer when insert performance is critical and load factor stays below 50%
- **Robin Hood**: Prefer when lookup misses are common or operating at higher load factors
