package gee

import (
	"strings"
)

type node struct {
	pattern 	string	// 待匹配的路由
	part 		string  // 路由中的一部分
	children 	[]*node	// 子节点
	isWild		bool 	// 若节点所表示的是动态匹配 则该字段为true
}

// 找到第一个与part匹配的节点 用于插入操作
func (n *node) matchChild(part string) *node {
	for _,child:=range n.children {
		if child.part==part || child.isWild {
			return child
		}
	}
	return nil
}
// 找到所有与part匹配的子节点 用于查询操作
func (n *node) matchChildren(part string) []*node {
	nodes:=make([]*node,0)
	for _,child:=range n.children {
		if child.part==part || child.isWild {
			nodes=append(nodes,child)
		}
	}
	return nodes
}

// 插入操作：递归查询每一层的节点
func (n *node) insert(pattern string,parts []string,height int){
	if len(parts)==height {
		n.pattern=pattern
		return
	}
	part:=parts[height]
	child:=n.matchChild(part)	// 查找当前这一层第一个与part匹配的节点
	// 若无则新建一个节点
	if child==nil {
		child=&node{part:part,isWild:part[0]==':'||part[0]=='*'}
		n.children=append(n.children,child)
	}
	child.insert(pattern,parts,height+1)
}

// 查询操作
func (n *node) search(parts []string,height int) *node {
	// 匹配成功 或者 匹配到“*” 就返回
	// URL中的“*”必须是URL的最后一个匹配节点
	if len(parts)==height || strings.HasPrefix(n.part,"*") {
		// 只有当前节点是某一URL的尾节点 pattern才有值
		if n.pattern==""{
			return nil
		}
		return n
	}
	part:=parts[height]
	// 这里遍历多个子节点是因为可能有这种情况：
	// /gee/go 既可以匹配到/gee/go 又可以匹配到/gee/:lang(动态路由)
	children:=n.matchChildren(part)
	for _,child:=range children {
		result:=child.search(parts,height+1)
		if result!=nil {
			return result
		}
	}
	return nil
}