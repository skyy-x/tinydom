// Package tinydom 实现了一个简单的XML的DOM树构造工具.
package tinydom

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"unicode/utf8"
	"container/list"
	"os"
)

// XMLAttribute 是一个元素的属性的接口.
//
// 这是一份关于属性的注释.
type XMLAttribute interface {
	Name() string
	Value() string
	SetValue(string)
}

// XMLNode 定义了XML所有节点的基础设施，提供了基本的元素遍历、增删等操作,也提供了逆向转换能力.
type XMLNode interface {
	ToElement() XMLElement
	ToText() XMLText
	ToComment() XMLComment
	ToDocument() XMLDocument
	ToProcInst() XMLProcInst
	ToDirective() XMLDirective

	Value() string
	SetValue(newValue string)

	Document() XMLDocument

	NoChildren() bool
	Parent() XMLNode
	FirstChild() XMLNode
	LastChild() XMLNode
	Prev() XMLNode
	Next() XMLNode
	FirstChildElement(name string) XMLElement
	LastChildElement(name string) XMLElement
	PrevElement(name string) XMLElement
	NextElement(name string) XMLElement

	InsertBack(node XMLNode) XMLNode
	InsertFront(node XMLNode) XMLNode
	InsertEndChild(node XMLNode) XMLNode
	InsertFirstChild(node XMLNode) XMLNode

	InsertElementBack(name string) XMLElement
	InsertElementFront(name string) XMLElement
	InsertElementEndChild(name string) XMLElement
	InsertElementFirstChild(name string) XMLElement

	DeleteChildren()
	DeleteChild(node XMLNode)

	Split() XMLNode

	Accept(visitor XMLVisitor) bool

	// 被迫入侵的接口
	insertBeforeChild(beforeThis XMLNode, addThis XMLNode) XMLNode
	insertAfterChild(afterThis XMLNode, addThis XMLNode) XMLNode
	setParent(node XMLNode)
	setPrev(node XMLNode)
	setNext(node XMLNode)
	setDocument(doc XMLDocument)
	//impl() XMLNode

	unlink(child XMLNode)
}

// XMLElement  提供了访问XML基本节点元素的能力
//
// Name、SetName其实是Value和SetValue的别名，目的是为了使得接口更加符合直观理解。
//
// Text、SetText的作用是设置<node>与</node>之间的文字，虽然文字都是有XMLText对象来承载的，但是通常来说直接在XMLElement中访问会更加方便。
//
// FindAttribute和ForeachAttribute分别用于查找特定的XML节点的属性和遍历XML属性列表。
//
// Attribute、SetAttribute、DeleteAttribute用于读取和删除属性。
type XMLElement interface {
	XMLNode

	Name() string
	SetName(name string)

	FindAttribute(name string) XMLAttribute
	ForeachAttribute(callback func(attribute XMLAttribute) int) int

	AttributeCount() int
	Attribute(name string, def string) string
	SetAttribute(name string, value string) XMLAttribute
	DeleteAttribute(name string) XMLAttribute
	ClearAttributes()

	Text() string
	SetText(text string)
}

// XMLText 提供了对XML元素间文本的封装
type XMLText interface {
	XMLNode
	SetCDATA(isCData bool)
	CDATA() bool
}

// XMLComment 提供了对注释的封装
type XMLComment interface {
	XMLNode
	Comment() string
	SetComment(string)
}

// XMLProcInst 常用于表达XML处理指令,类似:<?xml version="1.0" encoding="UTF-8"?>
type XMLProcInst interface {
	XMLNode
	Target() string
	Instruction() string
}

// XMLDirective 用于表达`<!`与`>`之间的部分，一般为DTD
type XMLDirective interface {
	XMLNode
}

// XMLDocument 用于表达一个XML文档,这是整个XML文档的根
type XMLDocument interface {
	XMLNode
}

// XMLVisitor XML文档访问器,常用于遍历文档或者格式化输出XML文档
type XMLVisitor interface {
	VisitEnterDocument(XMLDocument) bool
	VisitExitDocument(XMLDocument) bool

	VisitEnterElement(XMLElement) bool
	VisitExitElement(XMLElement) bool

	VisitProcInst(XMLProcInst) bool
	VisitText(XMLText) bool
	VisitComment(XMLComment) bool
	VisitDirective(XMLDirective) bool
}

// XMLHandle XML文档处理器,其主要
type XMLHandle interface {
	Parent() XMLHandle
	FirstChild() XMLHandle
	LastChild() XMLHandle
	Prev() XMLHandle
	Next() XMLHandle
	FirstChildElement(name string) XMLHandle
	LastChildElement(name string) XMLHandle
	PrevElement(name string) XMLHandle
	NextElement(name string) XMLHandle

	ToNode() XMLNode
	ToElement() XMLElement
	ToText() XMLText
	ToComment() XMLComment
	ToDocument() XMLDocument
	ToProcInst() XMLProcInst
	ToDirective() XMLDirective
}

// =========================================================

type xmlAttributeImpl struct {
	name  string
	value string
}

func (a *xmlAttributeImpl) Name() string {
	return a.name
}

func (a *xmlAttributeImpl) Value() string {
	return a.value
}

func (a *xmlAttributeImpl) SetValue(newValue string) {
	a.value = newValue
}

// ==================================================================

type xmlNodeImpl struct {
	implobj XMLNode

	document XMLDocument
	parent   XMLNode
	value    string

	firstChild XMLNode
	lastChild  XMLNode

	prev XMLNode
	next XMLNode
}

func (n *xmlNodeImpl) setParent(node XMLNode) {
	n.parent = node
}

func (n *xmlNodeImpl) setPrev(node XMLNode) {
	n.prev = node
}

func (n *xmlNodeImpl) setNext(node XMLNode) {
	n.next = node
}

func (n *xmlNodeImpl) setDocument(doc XMLDocument) {
	n.document = doc
}

func (n *xmlNodeImpl) ToElement() XMLElement {
	return nil
}

func (n *xmlNodeImpl) ToText() XMLText {
	return nil
}

func (n *xmlNodeImpl) ToComment() XMLComment {
	return nil
}

func (n *xmlNodeImpl) ToDocument() XMLDocument {
	return nil
}

func (n *xmlNodeImpl) ToProcInst() XMLProcInst {
	return nil
}

func (n *xmlNodeImpl) ToDirective() XMLDirective {
	return nil
}

func (n *xmlNodeImpl) Value() string {
	return n.value
}

func (n *xmlNodeImpl) SetValue(newValue string) {
	n.value = newValue
}

func (n *xmlNodeImpl) Document() XMLDocument {
	return n.document
}

func (n *xmlNodeImpl) Parent() XMLNode {
	return n.parent
}

func (n *xmlNodeImpl) NoChildren() bool {
	return nil == n.firstChild
}

func (n *xmlNodeImpl) FirstChild() XMLNode {
	return n.firstChild
}

func (n *xmlNodeImpl) LastChild() XMLNode {
	return n.lastChild
}

func (n *xmlNodeImpl) Prev() XMLNode {
	return n.prev
}

func (n *xmlNodeImpl) Next() XMLNode {
	return n.next
}

func (n *xmlNodeImpl) FirstChildElement(name string) XMLElement {
	for item := n.firstChild; nil != item; item = item.Next() {
		elem := item.ToElement()
		if nil == elem {
			continue
		}

		if ("" == name) || (elem.Name() == name) {
			return elem
		}
	}

	return nil
}

func (n *xmlNodeImpl) LastChildElement(name string) XMLElement {

	for item := n.lastChild; nil != item; item = item.Prev() {
		elem := item.ToElement()
		if nil == elem {
			continue
		}

		if ("" == name) || (elem.Name() == name) {
			return elem
		}
	}

	return nil
}

func (n *xmlNodeImpl) PrevElement(name string) XMLElement {
	for item := n.prev; nil != item; item = item.Prev() {
		elem := item.ToElement()
		if nil == elem {
			continue
		}

		if ("" == name) || (elem.Name() == name) {
			return elem
		}
	}

	return nil
}

func (n *xmlNodeImpl) NextElement(name string) XMLElement {

	for item := n.next; nil != item; item = item.Next() {
		elem := item.ToElement()
		if nil == elem {
			continue
		}

		if ("" == name) || (elem.Name() == name) {
			return elem
		}
	}

	return nil
}

func (n *xmlNodeImpl) Split() XMLNode {

	if nil != n.parent {
		n.parent.unlink(n.implobj)
	}

	return n.implobj
}

func (n *xmlNodeImpl) unlink(child XMLNode) {
	//if child.impl() == n.firstChild {
	if child == n.firstChild {
		n.firstChild = n.firstChild.Next()
	}

	//if child.impl() == n.lastChild {
	if child == n.lastChild {
		n.lastChild = n.lastChild.Prev()
	}

	if nil != child.Prev() {
		child.Prev().setNext(child.Next())
	}

	if nil != child.Next() {
		child.Next().setPrev(child.Prev())
	}

	child.setParent(nil)

	child.setDocument(nil)
}

func (n *xmlNodeImpl) InsertEndChild(addThis XMLNode) XMLNode {
	addThis.Split()

	if nil != n.lastChild {
		n.lastChild.setNext(addThis)
		addThis.setPrev(n.lastChild)
		n.lastChild = addThis
		addThis.setNext(nil)
	} else {
		n.firstChild = addThis
		n.lastChild = addThis

		addThis.setPrev(nil)
		addThis.setNext(nil)
	}

	addThis.setParent(n.implobj)
	addThis.setDocument(n.document)
	return addThis
}

func (n *xmlNodeImpl) InsertFirstChild(addThis XMLNode) XMLNode {
	addThis.Split()

	if nil != n.firstChild {
		n.firstChild.setPrev(addThis)
		addThis.setNext(n.firstChild)
		n.firstChild = addThis
		addThis.setPrev(nil)
	} else {
		n.firstChild = addThis
		n.lastChild = addThis

		addThis.setPrev(nil)
		addThis.setNext(nil)
	}

	addThis.setParent(n.implobj)
	addThis.setDocument(n.document)
	return addThis
}

func (n *xmlNodeImpl) insertAfterChild(afterThis XMLNode, addThis XMLNode) XMLNode {

	// if afterThis.Parent() != a.implobj {
	// return nil
	// }

	if afterThis.Next() == nil {
		return n.InsertEndChild(addThis)
	}

	addThis.Split()

	addThis.setPrev(afterThis)
	addThis.setNext(afterThis.Next())
	afterThis.Next().setPrev(addThis)
	afterThis.setNext(addThis)
	addThis.setParent(n.implobj)
	addThis.setDocument(n.document)

	return addThis
}

func (n *xmlNodeImpl) insertBeforeChild(beforeThis XMLNode, addThis XMLNode) XMLNode {

	// if beforeThis.Parent() != a.implobj {
	// return nil
	// }

	if beforeThis.Prev() == nil {
		return n.InsertFirstChild(addThis)
	}

	addThis.Split()

	addThis.setPrev(beforeThis.Prev())
	addThis.setNext(beforeThis)
	beforeThis.Prev().setNext(addThis)
	beforeThis.setPrev(addThis)
	addThis.setParent(n.implobj)
	addThis.setDocument(n.document)

	return addThis
}

func (n *xmlNodeImpl) InsertBack(addThis XMLNode) XMLNode {
	if nil == n.parent {
		return nil
	}

	return n.parent.insertAfterChild(n.implobj, addThis)
}

func (n *xmlNodeImpl) InsertFront(addThis XMLNode) XMLNode {
	if nil == n.parent {
		return nil
	}

	return n.parent.insertBeforeChild(n.implobj, addThis)
}

func (n *xmlNodeImpl) InsertElementFront(name string) XMLElement {
	return n.InsertFront(NewElement(name)).ToElement()
}

func (n *xmlNodeImpl) InsertElementBack(name string) XMLElement {
	return n.InsertBack(NewElement(name)).ToElement()
}

func (n *xmlNodeImpl) InsertElementEndChild(name string) XMLElement {
	return n.InsertEndChild(NewElement(name)).ToElement()
}

func (n *xmlNodeImpl) InsertElementFirstChild(name string) XMLElement {
	return n.InsertFirstChild(NewElement(name)).ToElement()
}

func (n *xmlNodeImpl) DeleteChildren() {
	for nil != n.firstChild {
		n.DeleteChild(n.firstChild)
	}

	n.firstChild = nil
	n.lastChild = nil
}

func (n *xmlNodeImpl) DeleteChild(node XMLNode) {
	n.unlink(node)
}

//func (n *xmlNodeImpl) Accept(visitor XMLVisitor) bool {
//	return n.implobj.Accept(visitor)
//}

// ------------------------------------------------------------------

type xmlElementImpl struct {
	xmlNodeImpl

	// rootAttribute XMLAttribute
	attrlist *list.List
	attrsmap map[string]*list.Element
}

func (e *xmlElementImpl) ToElement() XMLElement {
	return e
}

func (e *xmlElementImpl) Accept(visitor XMLVisitor) bool {

	if visitor.VisitEnterElement(e) {
		for node := e.FirstChild(); nil != node; node = node.Next() {
			if !node.Accept(visitor) {
				break
			}
		}
	}

	return visitor.VisitExitElement(e)
}

func (e *xmlElementImpl) Name() string {
	return e.Value()
}

func (e *xmlElementImpl) SetName(name string) {
	e.SetValue(name)
}

func (e *xmlElementImpl) FindAttribute(name string) XMLAttribute {
	elem, ok := e.attrsmap[name]
	if !ok {
		return nil
	}

	return elem.Value.(*xmlAttributeImpl)
}

func (e *xmlElementImpl) AttributeCount() int {
	return len(e.attrsmap)
}

func (e *xmlElementImpl) Attribute(name string, def string) string {
	attr, ok := e.attrsmap[name]
	if !ok {
		return def
	}

	return attr.Value.(*xmlAttributeImpl).Value()
}

func (e *xmlElementImpl) SetAttribute(name string, value string) XMLAttribute {
	elem, ok := e.attrsmap[name]
	if ok {
		elem.Value.(*xmlAttributeImpl).SetValue(value)
		return elem.Value.(*xmlAttributeImpl)
	}

	attr := newAttribute(name, value)
	e.attrsmap[name] = e.attrlist.PushBack(attr)
	return attr
}

func (e *xmlElementImpl) DeleteAttribute(name string) XMLAttribute {
	elem, ok := e.attrsmap[name]
	if !ok {
		return nil
	}

	attr := elem.Value.(*xmlAttributeImpl)

	e.attrlist.Remove(elem)
	delete(e.attrsmap, name)
	return attr
}

func (e *xmlElementImpl) Text() string {
	if text := e.FirstChild(); (nil != text) && (nil != text.ToText()) {
		return text.Value()
	}

	return ""
}

func (e *xmlElementImpl) SetText(inText string) {
	if node := e.FirstChild(); (nil != node) && (nil != node.ToText()) {
		node.SetValue(inText)
	} else {
		theText := NewText(inText)
		e.InsertFirstChild(theText)
	}
}

func (e *xmlElementImpl) ForeachAttribute(callback func(attribute XMLAttribute) int) int {
	for elem := e.attrlist.Front(); nil != elem; elem = elem.Next() {
		if ret := callback(elem.Value.(*xmlAttributeImpl)); 0 != ret {
			return ret
		}
	}

	return 0
}

func (e *xmlElementImpl) ClearAttributes() {
	e.attrlist = list.New()
	e.attrsmap = make(map[string]*list.Element)
}

// ------------------------------------------------------------------

type xmlCommentImpl struct {
	xmlNodeImpl
}

func (c *xmlCommentImpl) ToComment() XMLComment {
	return c
}

func (c *xmlCommentImpl) Comment() string {
	return c.value
}

func (c *xmlCommentImpl) SetComment(newComment string) {
	c.value = newComment
}

func (c *xmlCommentImpl) Accept(visitor XMLVisitor) bool {
	return visitor.VisitComment(c)
}

// ------------------------------------------------------------------

type xmlProcInstImpl struct {
	xmlNodeImpl
	instruction string
}

func (p *xmlProcInstImpl) ToProcInst() XMLProcInst {
	return p
}

func (p *xmlProcInstImpl) Accept(visitor XMLVisitor) bool {
	return visitor.VisitProcInst(p)
}

func (p *xmlProcInstImpl) Target() string {
	return p.value
}

func (p *xmlProcInstImpl) Instruction() string {
	return p.instruction
}

// ------------------------------------------------------------------

type xmlDocumentImpl struct {
	xmlNodeImpl
}

func (d *xmlDocumentImpl) ToDocument() XMLDocument {
	return d
}

func (d *xmlDocumentImpl) Accept(visitor XMLVisitor) bool {

	if visitor.VisitEnterDocument(d) {
		for node := d.FirstChild(); nil != node; node = node.Next() {
			if !node.Accept(visitor) {
				break
			}
		}
	}

	return visitor.VisitExitDocument(d)
}

// ------------------------------------------------------------------

type xmlTextImpl struct {
	xmlNodeImpl
	cdata bool
}

func (t *xmlTextImpl) ToText() XMLText {
	return t
}
func (t *xmlTextImpl) Accept(visitor XMLVisitor) bool {
	return visitor.VisitText(t)
}
func (t *xmlTextImpl) SetCDATA(isCData bool) {
	t.cdata = isCData
}
func (t *xmlTextImpl) CDATA() bool {
	return t.cdata
}

// ------------------------------------------------------------------

type xmlDirectiveImpl struct {
	xmlNodeImpl
}

func (d *xmlDirectiveImpl) ToDirective() XMLDirective {
	return d
}

func (d *xmlDirectiveImpl) Accept(visitor XMLVisitor) bool {
	return visitor.VisitDirective(d)
}

// ------------------------------------------------------------------

// NewText 创建一个新的XMLText对象
func NewText(text string) XMLText {
	node := new(xmlTextImpl)
	node.implobj = node
	node.value = text
	return node
}

// NewComment 创建一个新的XMLComment对象
func NewComment(comment string) XMLComment {
	node := new(xmlCommentImpl)
	node.implobj = node
	node.value = comment
	return node
}

// NewElement 创建一个新的XMLElement对象
func NewElement(name string) XMLElement {
	node := new(xmlElementImpl)
	node.implobj = node
	node.value = name
	node.attrsmap = make(map[string]*list.Element)
	node.attrlist = list.New()
	return node
}

// NewProcInst 创建一个新的XMLProcInst对象
func NewProcInst(target string, inst string) XMLProcInst {
	node := new(xmlProcInstImpl)
	node.implobj = node
	node.value = target
	node.instruction = inst
	return node
}

// NewDirective 创建一个新的XMLDirective对象
func NewDirective(directive string) XMLDirective {
	node := new(xmlDirectiveImpl)
	node.implobj = node
	node.value = directive
	return node
}

// newAttribute 创建一个新的XMLAttribute对象.
// name和value分别用于指定属性的名称和值
func newAttribute(name string, value string) *xmlAttributeImpl {
	attr := new(xmlAttributeImpl)
	attr.name = name
	attr.value = value
	return attr
}

// NewDocument 创建一个全新的XMLDocument对象
func NewDocument() XMLDocument {
	doc := new(xmlDocumentImpl)
	doc.implobj = doc
	doc.document = doc
	return doc
}

type context struct {
	doc           XMLDocument
	parent        XMLNode
	rootElemExist bool
}

func handleStartElement(startElement xml.StartElement, ctx *context) error {
	//startElement := token.(xml.StartElement)

	// 一个XML文档只允许有唯一一个根节点
	if ctx.doc == ctx.parent {
		if ctx.rootElemExist {
			return errors.New("Root element has been exist:" + startElement.Name.Local)
		}

		// 标记一下根节点已经存在了
		ctx.rootElemExist = true
	}

	node := NewElement(startElement.Name.Local)
	for _, item := range startElement.Attr {
		if nil != node.FindAttribute(item.Name.Local) {
			return errors.New("Attributes have the same name:" + item.Name.Local)
		}
		node.SetAttribute(item.Name.Local, item.Value)
	}
	ctx.parent.InsertEndChild(node)
	ctx.parent = node

	return nil
}

func handleCharData(charData xml.CharData, ctx *context) error {
	shortCharData := bytes.TrimSpace(charData)
	if (nil != shortCharData) && (len(shortCharData) > 0) {
		if ctx.doc == ctx.parent {
			return errors.New("Text should be in the element")
		}

		node := NewText(string(charData))
		node.SetCDATA(true)
		ctx.parent.InsertEndChild(node)
	}

	return nil
}

// LoadDocument 从rd流中读取XML码流并构建成XMLDocument对象
func LoadDocument(rd io.Reader) (XMLDocument, error) {

	// 创建一个context
	ctx := new(context)
	ctx.doc = NewDocument()
	ctx.parent = ctx.doc
	ctx.rootElemExist = false

	// 创建一个decoder
	decoder := xml.NewDecoder(rd)
	var token xml.Token
	var err error

	for token, err = decoder.Token(); nil == err; token, err = decoder.Token() {
		switch token.(type) {
		case xml.StartElement:
			err := handleStartElement(token.(xml.StartElement), ctx)
			if nil != err {
				return nil, err
			}
		case xml.EndElement:
			ctx.parent = ctx.parent.Parent()
		case xml.Comment:
			ctx.parent.InsertEndChild(NewComment(string(token.(xml.Comment))))
		case xml.Directive:
			ctx.parent.InsertEndChild(NewDirective(string(token.(xml.Directive))))
		case xml.ProcInst:
			procInst := token.(xml.ProcInst)
			ctx.parent.InsertEndChild(NewProcInst(procInst.Target, string(procInst.Inst)))
		case xml.CharData:
			if err := handleCharData(token.(xml.CharData), ctx); nil != err {
				return nil, err
			}
		default:
			return nil, errors.New("Unsupported token type")
		}
	}

	if (nil == err) || (io.EOF == err) {
		// 不能是空文档
		if nil == ctx.doc.FirstChildElement("") {
			return nil, errors.New("XML document missing the root element")
		}

		return ctx.doc, nil
	}

	return nil, err
}

func LoadDocumentFromFile(name string) (XMLDocument, error) {
	file, err := os.Open(name)
	if nil != err {
		return nil, err
	}
	defer file.Close()

	return LoadDocument(file)
}

// SaveDocumentToFile Print the xml-dom objects to the writer.
func SaveDocument(doc XMLDocument, writer io.Writer, options PrintOptions) error {
	doc.Accept(NewSimplePrinter(writer, options))
	return nil
}

// SaveDocumentToFile Print the xml-dom objects to the file.
func SaveDocumentToFile(doc XMLDocument, name string, options PrintOptions) error {
	file, err := os.Create(name)
	if nil != err {
		return err
	}
	defer file.Close()

	doc.Accept(NewSimplePrinter(file, options))
	return nil
}

// DefaultVisitor 这个类的目的是简化编写定制扫描的visitor,使得我们不需要定制XMLVisitor的所有接口
type DefaultVisitor struct {
	EnterDocument func(XMLDocument) bool
	ExitDocument  func(XMLDocument) bool
	EnterElement  func(XMLElement) bool
	ExitElement   func(XMLElement) bool
	ProcInst      func(XMLProcInst) bool
	Text          func(XMLText) bool
	Comment       func(XMLComment) bool
	Directive     func(XMLDirective) bool
}

// VisitEnterDocument is the default implement of XMLVisitor
func (v *DefaultVisitor) VisitEnterDocument(doc XMLDocument) bool {
	if nil == v.EnterDocument {
		return true
	}

	return v.EnterDocument(doc)
}

// VisitExitDocument is the default implement of XMLVisitor
func (v *DefaultVisitor) VisitExitDocument(doc XMLDocument) bool {
	if nil == v.ExitDocument {
		return true
	}

	return v.ExitDocument(doc)
}

// VisitEnterElement is the default implement of XMLVisitor
func (v *DefaultVisitor) VisitEnterElement(elem XMLElement) bool {
	if nil == v.EnterElement {
		return true
	}

	return v.EnterElement(elem)
}

// VisitExitElement is the default implement of XMLVisitor
func (v *DefaultVisitor) VisitExitElement(elem XMLElement) bool {
	if nil == v.ExitElement {
		return true
	}

	return v.ExitElement(elem)
}

// VisitProcInst is the default implement of XMLVisitor
func (v *DefaultVisitor) VisitProcInst(pi XMLProcInst) bool {
	if nil == v.ProcInst {
		return true
	}

	return v.ProcInst(pi)
}

// VisitText is the default implement of XMLVisitor
func (v *DefaultVisitor) VisitText(text XMLText) bool {
	if nil == v.Text {
		return true
	}

	return v.Text(text)
}

// VisitComment is the default implement of XMLVisitor
func (v *DefaultVisitor) VisitComment(c XMLComment) bool {
	if nil == v.Comment {
		return true
	}

	return v.Comment(c)
}

// VisitDirective is the default implement of XMLVisitor
func (v *DefaultVisitor) VisitDirective(d XMLDirective) bool {
	if nil == v.Directive {
		return true
	}

	return v.Directive(d)
}

// ------------------------------------------------------------------
type xmlSimplePrinter struct {
	writer      io.Writer    // 输出目的地
	options     PrintOptions // 格式化选项
	level       int          // 用于缩进时指定缩进级别
	firstPrint  bool         // 是否首次输出
	indentBytes []byte       // 索引字符流
	lineHold    bool         // 暂停换行
}

// PrintOptions    打印选项,用于NewSimplePrinter函数,用于控制输出的XML内容的样式
type PrintOptions struct {
	Indent        []byte // 缩进前缀,只允许填写tab或者空白,如果Indent长度为0表示折行但是不缩进,如果Indent为null表示不折行
	TextWrapWidth int    // 超过多长才强制换行
}

var (
	// PrintPretty  预制的打印选项,采用4个空格缩进
	PrintPretty = PrintOptions{Indent: []byte("    "), TextWrapWidth: 200}

	// PrintStream 流式打印选项,不缩进,不换行,节省流量
	PrintStream = PrintOptions{}
)

// NewSimplePrinter 创建一个简单XML文档输出函数
func NewSimplePrinter(writer io.Writer, options PrintOptions) XMLVisitor {
	visitor := new(xmlSimplePrinter)
	visitor.writer = writer
	visitor.options = options
	visitor.level = 0
	visitor.firstPrint = true
	return visitor
}

func (p *xmlSimplePrinter) indentSpace() {
	if nil != p.options.Indent {
		if len(p.options.Indent) >= 0 {
			if !p.firstPrint {
				p.writer.Write([]byte("\n"))
			}
		}
	}

	for i := 0; i < p.level; i++ {
		p.writer.Write(p.options.Indent)
	}

	p.firstPrint = false
}

func (p *xmlSimplePrinter) VisitEnterDocument(node XMLDocument) bool {
	return true
}

func (p *xmlSimplePrinter) VisitExitDocument(node XMLDocument) bool {
	return true
}

func (p *xmlSimplePrinter) VisitEnterElement(node XMLElement) bool {
	p.indentSpace()
	p.level++

	p.writer.Write([]byte("<"))
	p.writer.Write([]byte(node.Name()))

	node.ForeachAttribute(func(attribute XMLAttribute) int {
		p.writer.Write([]byte(` `))
		p.writer.Write([]byte(attribute.Name()))
		p.writer.Write([]byte(`="`))
		EscapeAttribute(p.writer, []byte(attribute.Value()))
		p.writer.Write([]byte(`"`))
		return 0
	})

	if node.NoChildren() {
		p.level--
		p.writer.Write([]byte("/>"))
		return true
	}

	p.writer.Write([]byte(">"))
	return true
}

func (p *xmlSimplePrinter) VisitExitElement(node XMLElement) bool {
	if node.NoChildren() {
		return true
	}

	p.level--
	p.indentSpace()
	p.writer.Write([]byte("</"))
	p.writer.Write([]byte(node.Name()))
	p.writer.Write([]byte(">"))
	return true
}

func (p *xmlSimplePrinter) VisitProcInst(node XMLProcInst) bool {
	p.indentSpace()
	p.writer.Write([]byte("<?"))
	p.writer.Write([]byte(node.Target()))
	p.writer.Write([]byte(" "))
	p.writer.Write([]byte(node.Instruction()))
	p.writer.Write([]byte("?>"))
	return true
}

func (p *xmlSimplePrinter) VisitText(node XMLText) bool {
	p.indentSpace()
	if node.CDATA() {
		p.writer.Write([]byte("<![CDATA["))
		p.writer.Write([]byte(node.Value()))
		p.writer.Write([]byte("]]>"))
		return true
	}

	EscapeText(p.writer, []byte(node.Value()))
	return true
}

func (p *xmlSimplePrinter) VisitComment(node XMLComment) bool {
	p.indentSpace()
	p.writer.Write([]byte("<!--"))
	p.writer.Write([]byte(node.Value()))
	p.writer.Write([]byte("-->"))
	return true
}

func (p *xmlSimplePrinter) VisitDirective(node XMLDirective) bool {
	p.indentSpace()
	p.writer.Write([]byte("<!"))
	EscapeText(p.writer, []byte(node.Value()))
	p.writer.Write([]byte(">"))
	return true
}

// ------------------------------------------------------------------

type xmlHandleImpl struct {
	node XMLNode
}

// NewHandle 创建一个新的XMLHandle对象
func NewHandle(node XMLNode) XMLHandle {
	handle := new(xmlHandleImpl)
	handle.node = node
	return handle
}

func (h *xmlHandleImpl) Parent() XMLHandle {
	if nil == h.node {
		return h
	}

	return NewHandle(h.node.Parent())
}

func (h *xmlHandleImpl) FirstChild() XMLHandle {
	if nil == h.node {
		return h
	}

	return NewHandle(h.node.FirstChild())
}

func (h *xmlHandleImpl) LastChild() XMLHandle {
	if nil == h.node {
		return h
	}

	return NewHandle(h.node.LastChild())
}

func (h *xmlHandleImpl) Prev() XMLHandle {
	if nil == h.node {
		return h
	}

	return NewHandle(h.node.Prev())
}

func (h *xmlHandleImpl) Next() XMLHandle {
	if nil == h.node {
		return h
	}

	return NewHandle(h.node.Next())
}

func (h *xmlHandleImpl) FirstChildElement(name string) XMLHandle {
	if nil == h.node {
		return h
	}

	return NewHandle(h.node.FirstChildElement(name))
}

func (h *xmlHandleImpl) LastChildElement(name string) XMLHandle {
	if nil == h.node {
		return h
	}

	return NewHandle(h.node.LastChildElement(name))
}

func (h *xmlHandleImpl) PrevElement(name string) XMLHandle {
	if nil == h.node {
		return h
	}

	return NewHandle(h.node.PrevElement(name))
}

func (h *xmlHandleImpl) NextElement(name string) XMLHandle {
	if nil == h.node {
		return h
	}

	return NewHandle(h.node.NextElement(name))
}

func (h *xmlHandleImpl) ToNode() XMLNode {
	return h.node
}

func (h *xmlHandleImpl) ToElement() XMLElement {
	if nil == h.node {
		return nil
	}

	return h.node.ToElement()
}

func (h *xmlHandleImpl) ToText() XMLText {
	if nil == h.node {
		return nil
	}

	return h.node.ToText()
}

func (h *xmlHandleImpl) ToComment() XMLComment {
	if nil == h.node {
		return nil
	}

	return h.node.ToComment()
}

func (h *xmlHandleImpl) ToDocument() XMLDocument {
	if nil == h.node {
		return nil
	}

	return h.node.ToDocument()
}

func (h *xmlHandleImpl) ToProcInst() XMLProcInst {
	if nil == h.node {
		return nil
	}

	return h.node.ToProcInst()
}

func (h *xmlHandleImpl) ToDirective() XMLDirective {
	if nil == h.node {
		return nil
	}

	return h.node.ToDirective()
}

// isInCharacterRange 这个函数是直接从xml包里面拷贝出来的
// Decide whether the given rune is in the XML Character Range, per
// the Char production of http:// www.xml.com/axml/testaxml.htm,
// Section 2.2 Characters.
func isInCharacterRange(r rune) (inrange bool) {
	return r == 0x09 ||
		r == 0x0A ||
		r == 0x0D ||
		r >= 0x20 && r <= 0xDF77 ||
		r >= 0xE000 && r <= 0xFFFD ||
		r >= 0x10000 && r <= 0x10FFFF
}

// 最简洁的字符
// 字符    属性    文本    转义
// &       no     no     &amp;
// <       no     no     &lt;
// "       no     yes    &quot;
// \n      no     yes    &#xA;
// \r      no     yes    &#xD;
// '       yes    yes    &apos;
// >       yes    yes    &gt;
var (
	escAmps = []byte("&amp;")
	escLt   = []byte("&lt;")
	escQuot = []byte("&quot;")
	escNl   = []byte("&#xA;")
	escCr   = []byte("&#xD;")
	escFFFD = []byte("\uFFFD") // Unicode replacement character
)

// EscapeAttribute 对XMLElement中的属性值进行转义,常用于自定义文档输出格式
func EscapeAttribute(w io.Writer, s []byte) error {
	var esc []byte
	last := 0
	for i := 0; i < len(s); {
		r, width := utf8.DecodeRune(s[i:])
		i += width
		switch r {
		case '&':
			esc = escAmps
		case '<':
			esc = escLt
		case '"':
			esc = escQuot
		case '\n':
			esc = escNl
		case '\r':
			esc = escCr
		default:
			if !isInCharacterRange(r) || (r == 0xFFFD && width == 1) {
				esc = escFFFD
				break
			}
			continue
		}
		if _, err := w.Write(s[last: i-width]); err != nil {
			return err
		}
		if _, err := w.Write(esc); err != nil {
			return err
		}
		last = i
	}
	if _, err := w.Write(s[last:]); err != nil {
		return err
	}
	return nil
}

// EscapeText 对文本内容进行转义,常用于自定义文档输出格式
func EscapeText(w io.Writer, s []byte) error {
	var esc []byte
	last := 0
	for i := 0; i < len(s); {
		r, width := utf8.DecodeRune(s[i:])
		i += width
		switch r {
		case '&':
			esc = escAmps
		case '<':
			esc = escLt
		default:
			if !isInCharacterRange(r) || (r == 0xFFFD && width == 1) {
				esc = escFFFD
				break
			}
			continue
		}
		if _, err := w.Write(s[last: i-width]); err != nil {
			return err
		}
		if _, err := w.Write(esc); err != nil {
			return err
		}
		last = i
	}
	if _, err := w.Write(s[last:]); err != nil {
		return err
	}
	return nil
}

// Version 查询版本信息
func Version() string {
	return "1.2.0"
}
