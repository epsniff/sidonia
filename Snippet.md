# Snippet


## Linear merging of posting lists 
eval := andEval(doc1, doc2) {
    if doc1 = doc2 {  
        emit(doc1)
    }
}

eval := orEval(doc1, doc2) {
    if doc1 != nil {
        emit(doc1)
    } else {
        emit(doc2)
    }
}

eval := andNot(doc1, doc2) {
    if doc2 == nil {  // We assume that if doc2 is nil, then doc1 is not nil
        emit(doc1)
    }
}

while p1[p1Idx] != nil and p2[p2Idx] != nil do 
  if docID(p1[p1Idx]) == docID(p2[p2Idx]) 
       eval( docID(p1[p1Idx], docID(p2[p2Idx]) 
       p1Idx++
       p2Idx++
  else if docID(p1[p1Idx]) < docID(p2[p2Idx])
       eval( docID(p1[p1Idx],nil) 
       p1Idx++
  else
       eval( nil, docID(p2[p2Idx]) 
       p2Idx++