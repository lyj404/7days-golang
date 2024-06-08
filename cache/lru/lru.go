package lru

import "container/list"

type Cache struct {
	maxBytes  int64                         // 允许使用的最大内存
	nbytes    int64                         // 当前已使用的内存
	ll        *list.List                    // 双向链表
	cache     map[string]*list.Element      // 键是字符串，值是双向链表中对应节点的指针
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数，可以为 nil
}

// 键值对 entry 是双向链表节点的数据类型，在链表中仍保存每个值对应的 key 的好处在于，淘汰队首节点时，需要用 key 从字典中删除对应的映射
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int // 用于返回值所占用的内存大小
}

// New 是Cache的构造方法，用来实例化Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	// 从字段中找到对应的双向链表的节点
	if ele, ok := c.cache[key]; ok {
		// 使用链表的 MoveToFront 方法将该元素移动到链表的前端
		c.ll.MoveToFront(ele)
		// // 从链表节点中提取值，ele.Value 已经知道是 *entry 类型
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 方法用于从缓存中移除最老的条目，即最近最少使用的条目
func (c *Cache) RemoveOldest() {
	// 获取双向链表中最后一个节点
	ele := c.ll.Back()
	if ele != nil {
		// c.ll.Remove(ele) 从链表中移除这个最老的节点
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// cache 映射中删除与 entry 的 key 相关联的条目，确保映射和链表同步
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	// 尝试从 cache 映射中获取与 key 相关联的双向链表节点 *list.Element。如果 key 存在，ok 将为 true
	if ele, ok := c.cache[key]; ok {
		// 使用链表的 MoveToFront 方法将该元素移动到链表的前端，表示这个键是最近访问的
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		// 更新缓存的总字节数 c.nbytes。如果替换了缓存中的值，需要调整字节数，增加新值的字节数减去旧值的字节数
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		// 将新的 value 赋值给 entry 结构体的 value 字段
		kv.value = value
	} else {
		// 使用 PushFront 方法将新的 entry（包含 key 和 value）添加到链表的前端
		ele := c.ll.PushFront(&entry{
			key:   key,
			value: value,
		})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// 如果设置了最大字节数 c.maxBytes 并且当前缓存的总字节数 c.nbytes 超过了这个限制，则删除最近最少使用的元素
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Len 返回缓存中条目的数量
func (c *Cache) Len() int {
	return c.ll.Len()
}
