package gee

import "strings"

type node struct {
	pattern  string  // 待匹配的路由，例如/p/:lang
	part     string  // 路由中的一部分，例如 :lang
	children []*node // 子节点，例如 [doc, hello, index]
	isWild   bool    // 是否精确匹配，part含有：或者*时为true
}

// 第一个匹配到的节点，用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		// child.part == part：表示当前子节点的部分与给定部分完全匹配，例如 /user 与 /user 完全匹配
		// child.isWild 表示当前子节点是通配节点，可以匹配任意部分，例如 /user/:id 中的 :id 就是一个通配节点，可以匹配任意字符串
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 匹配所有成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// 插入路由规则到路由树中
func (n *node) insert(pattern string, parts []string, height int) {
	// 如果已经匹配到最后一个部分，则将路由规则赋值给当前节点
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	// 获取当前层级对应的路由部分
	part := parts[height]
	// 查找是否存在对应的子节点，如果不存在则创建一个新的子节点
	child := n.matchChild(part)
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	// 递归插入下一个部分的路由规则
	child.insert(pattern, parts, height+1)
}

// 根据路由规则查找对应的节点
func (n *node) search(parts []string, height int) *node {
	// 如果已经匹配到最后一个部分或当前节点为通配节点，则返回当前节点
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	// 获取当前层级对应的路由部分
	part := parts[height]
	// 匹配所有成功的子节点
	children := n.matchChildren(part)

	// 遍历所有匹配成功的子节点
	for _, child := range children {
		// 递归查找下一个部分的路由规则
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}
