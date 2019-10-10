# Introduction 

This is a very raw implementation of an inverted index similar to lucene.  I'm using 
this as a playground to grow my own understanding of Information Retrieval solutions 
and algorithm used for full text searching databases like ElasticSearch.  


# Good reads:
[Index 1,600,000,000 Keys with Automata and Rust](https://blog.burntsushi.net/transducers/) - learning about FSTs 

https://www.elastic.co/guide/en/elasticsearch/reference/6.2/glossary.html

https://jingxuan.li/2018/08/25/Boolean-retrieval/ 
https://www.elastic.co/blog/apache-lucene-numeric-filters 
 - BKDtrees vs Enter Uwe Schindler (prefix encoding in FST)
   See https://github.com/blevesearch/bleve/blob/master/numeric/prefix_coded.go#L21 for an example of prefix encoding
https://medium.com/@nickgerleman/the-bkd-tree-da19cf9493fb


