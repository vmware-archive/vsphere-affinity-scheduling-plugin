# Kubernetes client-go's cache

Kubernetes [client-go](https://github.com/kubernetes/client-go) has a package `k8s.io/client-go/tools/cache`
which implements a client side cache mechanism. The cache is very powerful, basically it supports cache 
any objects based on what process function the client provides. In order to write an efficient client side
cache, a good understanding of what the cache does is important.

## New methods

NewInformer gives you a store and controller.

- NewInformer() (Store, Controller)
- NewIndexInformer() (Indexer, Controller)

NewSharedInformer gives you one (shared) informer interface which includes both a store and controller.
 
- NewSharedInformer() SharedInformer
- NewSharedIndexInformer() SharedIndexInformer

## Interfaces

### Store and Indexer

Store is generic interface that knows how to both retrieve and update objects.

- Retrieve is for the users to get cached objects, e.g. List, ListKeys, Get, GetByKey.
- Update methods are for Reflector to update the cache to reflect the latest change on server. e.g. Add, Update, Delete.

Indexer is an extension on Store to support index on fields other than the object key. It also supports
add more indexes to store. The index has to be added before the store has data in it. Index is useful when
you want to have fast label selector, etc.

- Indexes support includes methods like Index, IndexKeys, ByIndex
- Get indexes: AddIndexers, GetIndexers

#### UndeltaStore

It's a store with a PushFunc that pushes the whole state once the state of store changed.

#### ThreadSafeStore

ThreadSafeStore is in interface which mostly overlaps the Indexer. It shouldn't be. Anyway it's just an implementation
detail that supports Store and Index. The user shouldn't care about this.

### Controller and Reflector

Run the Controller then basically the reflector starts to watch server and update the store.

Reflector watches a specific resource and causes all changes to be refelcted in a given store.

### SharedInformer and SharedIndexInformer

SharedInformer includes
- Controller (it includes all 3 methods in controller)
- GetController()
- GetStore()
- AddEventHandler()

SharedIndexInformer includes
- SharedInformer
- AddIndexers()
- GetIndex()

## FIFO (Queue)

Queue as an interface is Store that supports Pop(PopFunc). I think it's pretty weird for Queue to include Store.
It is named as "queue" because it's used in Reflector to send updates to processor/handler to build cache's store.

- Store
- Pop(PopProcessFunc): Pop blocks until the process function finished processing popped the item.
- AddIfNotPresent(obj): Add item back to queue if failed to process.
- ...

so the interface methods include:
- Add/Update/Delete: for reflector's new watched object to add to the queue.
- Pop: to get an item from queue and apply to cache's store.
 
Both FIFO and DeltaFIFO are implementations of Queue. The difference is in FIFO you process the final state of an
object, while DeltaFIFO you want to process every change to an object. Also DeltaFIFO deltaFIFO deals with object
deletion, while FIFO doesn't.

Only DeltaFIFO is used in this package, probably because they want to deals with deletion. NewInformer uses
DeltaFIFO, the process function's input object type is Deltas in that case. Delta includes type [Add,Update,Delete]
and object itself.

