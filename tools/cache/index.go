/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Indexer is a storage interface that lets you list objects using multiple indexing functions
type Indexer interface {
	Store

	// Retrieve list of objects that match on the named indexing function
	// indexName是具体某个IndexFunc的名字
	// indexName索引类，obj是对象，计算obj在indexName索引类中的索引键，通过索引键把所有的对象取出来
	// 基本就是获取符合obj特征的所有对象，所谓的特征就是对象在索引类中的索引键, 一个典型的用法:
	// 获取在某个namespace下的所有对象, 这里NamespaceIndex就是一种具体的索引类
	// items, err := indexer.Index(NamespaceIndex, &metav1.ObjectMeta{Namespace: namespace})
	// 一句话总结: 传入索引类名称(索引函数名称), 获取所有符合 obj特征的对象
	Index(indexName string, obj interface{}) ([]interface{}, error)

	// IndexKeys returns the set of keys that match on the named indexing function.
	// indexKey是indexName索引类中一个索引键，函数返回indexKey指定的所有对象键
	// 这个对象键是Indexer内唯一的，在添加的时候会计算，后面讲具体Indexer实例的会讲解
	IndexKeys(indexName, indexKey string) ([]string, error)

	// ListIndexFuncValues returns the list of generated values of an Index func
	// 获取indexName索引类中的所有索引键
	ListIndexFuncValues(indexName string) []string

	// ByIndex lists object that match on the named indexing function with the exact key
	// 这个函数和Index类似，只是返回值不是对象键，而是所有对象
	ByIndex(indexName, indexKey string) ([]interface{}, error)

	// GetIndexer return the indexers
	GetIndexers() Indexers

	// AddIndexers adds more indexers to this store.  If you call this after you already have data
	// in the store, the results are undefined.
	// 添加Indexers，就是增加更多的索引分类
	AddIndexers(newIndexers Indexers) error
}

// IndexFunc knows how to provide an indexed value for an object.
// 计算索引的函数，传入对象，输出字符串索引，注意是数组
type IndexFunc func(obj interface{}) ([]string, error)

// IndexFuncToKeyFuncAdapter adapts an indexFunc to a keyFunc.  This is only useful if your index function returns
// unique values for every object.  This is conversion can create errors when more than one key is found.  You
// should prefer to make proper key and index functions.
func IndexFuncToKeyFuncAdapter(indexFunc IndexFunc) KeyFunc {
	return func(obj interface{}) (string, error) {
		indexKeys, err := indexFunc(obj)
		if err != nil {
			return "", err
		}
		if len(indexKeys) > 1 {
			return "", fmt.Errorf("too many keys: %v", indexKeys)
		}
		if len(indexKeys) == 0 {
			return "", fmt.Errorf("unexpected empty indexKeys")
		}
		return indexKeys[0], nil
	}
}

const (
	NamespaceIndex string = "namespace"
)

// MetaNamespaceIndexFunc is a default index function that indexes based on an object's namespace
func MetaNamespaceIndexFunc(obj interface{}) ([]string, error) {
	meta, err := meta.Accessor(obj)
	if err != nil {
		return []string{""}, fmt.Errorf("object has no meta: %v", err)
	}
	return []string{meta.GetNamespace()}, nil
}

// Index maps the indexed value to a set of keys in the store that match on that value
// 每种计算索引的方式会输出多个索引(数组), 而多个目标可能会算出相同索引，所以就有了这个类型
type Index map[string]sets.String

// Indexers maps a name to a IndexFunc
// 计算索引的函数有很多，用名字分类
type Indexers map[string]IndexFunc

// Indices maps a name to an Index
// 由于有多种计算索引的方式，那就又要按照计算索引的方式组织索引
type Indices map[string]Index
