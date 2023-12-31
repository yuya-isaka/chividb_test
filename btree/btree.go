package btree

import (
	"encoding/binary"
	"fmt"

	"github.com/yuya-isaka/chibidb/disk"
	"github.com/yuya-isaka/chibidb/pool"
)

// ======================================================================

/*

	If Meta
		MetaHeader (RootのページID保持) 		 ... 8 bytes
		MetaBody (空っぽ) 									... 4088 bytes

	If Node
		NodeHeader (Nodeの種類保持) 			... 8 bytes
		NodeBody (Leaf or Branch) 			 ... 4088 bytes
			If Leaf
				LeafHeader (Leafの前後のページID保持) 		 ... 16 bytes
				LeafBody (Slot) 												... 4072 bytes
					SlotHeader (スロット数と空きスペース保持) 			 ... 4 bytes
					SlotBody (KeyとValue) 										 ... 4068 bytes
			If Branch
				BranchHeader (右のページID保持) 					... 8 bytes
				BranchBody (Slot) 							 				... 4080 bytes
					SlotHeader (スロット数と空きスペース保持) 			 ... 4 bytes
					SlotBody (KeyとValue) 										 ... 4076 bytes

*/

// ======================================================================

type NodeType interface {
	xxxProtexted()
}

type nodetype string

func (n nodetype) xxxProtexted() {}

const (
	// Nodeの種類
	LeafNodeType   nodetype = "LEAF    " // 葉ノード、8 bytes
	BranchNodeType nodetype = "BRANCH  " // 枝ノード、8 bytes
)

// ======================================================================

// バイトスライスをdisk.PageIDに変換
func toPageID(b []byte) disk.PageID {
	if len(b) != 8 {
		return disk.InvalidPageID
	}
	// binary.LittleEndianで符号なし64ビット整数に変換
	return disk.PageID(binary.LittleEndian.Uint64(b))
}

// disk.PageIDを8bytesのバイトスライスに変換
func to8Bytes(i disk.PageID) []byte {
	b := make([]byte, 8)
	// binary.LittleEndianでバイトスライスに変換
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}

// uint16を2bytesのバイトスライスに変換
func to2Bytes(i uint16) []byte {
	b := make([]byte, 2)
	// binary.LittleEndianでバイトスライスに変換
	binary.LittleEndian.PutUint16(b, i)
	return b
}

// ======================================================================

// BTreeの始まりであるルートIDを保持する
type MetaHeader struct {
	rootID []byte // 8 bytes, disk.PageID
}

type Meta struct {
	header MetaHeader // 8 bytes
}

// rootIDも設定する
func NewMeta(page *pool.Page) (*Meta, error) {
	// 4096 bytes のページデータを取得
	pageData := page.GetPageData()
	if len(pageData) != disk.PageSize {
		return nil, fmt.Errorf("invalid page size: got %d, want %d", len(pageData), disk.PageSize)
	}

	// 新しいメタデータを作成し、ページデータからヘッダを抽出する
	meta := &Meta{
		header: MetaHeader{rootID: pageData[:8]}, // 8 bytes
	}

	return meta, nil
}

func (m *Meta) GetRootID() disk.PageID {
	return toPageID(m.header.rootID)
}

func (m *Meta) SetRootID(rootID disk.PageID) error {
	if rootID <= disk.InvalidPageID {
		return fmt.Errorf("invalid page id: got %d", rootID)
	}
	copy(m.header.rootID, to8Bytes(rootID))
	return nil
}

// ======================================================================

type NodeHeader struct {
	nodeType []byte // 8 bytes, LeafNodeType or BranchNodeType
}

type Node struct {
	header NodeHeader // 8 bytes
	body   []byte     // 4088 bytes
}

func NewNode(page *pool.Page) (*Node, error) {
	// 4096 bytes のページデータを取得
	pageData := page.GetPageData()
	if len(pageData) != disk.PageSize {
		return nil, fmt.Errorf("invalid page size: got %d, want %d", len(pageData), disk.PageSize)
	}

	// 新しいノードを作成し、ページデータからヘッダとボディを抽出する
	node := &Node{
		header: NodeHeader{nodeType: pageData[:8]}, // 8 bytes
		body:   pageData[8:],                       // 4088 bytes
	}

	return node, nil
}

func (n *Node) GetNodeType() NodeType {
	return nodetype(n.header.nodeType)
}

// ノードのヘッダーを初期化
func (n *Node) SetNodeType(nodeType NodeType) {
	if nt, ok := nodeType.(nodetype); ok {
		copy(n.header.nodeType, nt)
	}
}

// ======================================================================

type SlotHeader struct {
	numSlot   []byte // 2 bytes, uint16
	freeSpace []byte // 2 bytes, uint16
}

// 4072 bytes (Leafのbodyのサイズ) or 4080 bytes (Branchのbodyのサイズ)
type Slot struct {
	header SlotHeader // 4 bytes
	body   []byte     // 4068 bytes (Leaf) or 4076 bytes (Branch)
}

// 初期はスロット数0、空きスペースは全てのボディ
func (s *Slot) reset() {
	copy(s.header.numSlot, to2Bytes(0))
	copy(s.header.freeSpace, to2Bytes(uint16(len(s.body))))
}

// ======================================================================

type LeafHeader struct {
	prevID []byte // 8 bytes
	nextID []byte // 8 bytes
}

// 4088 bytes (Nodeのbodyのサイズ)
type Leaf struct {
	header LeafHeader // 16 bytes
	body   Slot       // 4072 bytes
}

func NewLeaf(node *Node) (*Leaf, error) {
	if nodetype(node.header.nodeType) != LeafNodeType {
		return nil, fmt.Errorf("invalid node type: got %s, want %s", node.header.nodeType, LeafNodeType)
	}

	// 4088 bytes のノードボディを取得
	nodeBody := node.body
	if len(nodeBody) != disk.PageSize-8 {
		return nil, fmt.Errorf("invalid page size: got %d, want %d", len(nodeBody), disk.PageSize-8)
	}

	leaf := &Leaf{
		// 16 bytes
		header: LeafHeader{
			prevID: nodeBody[:8],   // 8 bytes
			nextID: nodeBody[8:16], // 8 bytes
		},
		// 4072 bytes
		body: Slot{
			header: SlotHeader{
				numSlot:   nodeBody[16:18], // 2 bytes
				freeSpace: nodeBody[18:20], // 2 bytes
			},
			body: nodeBody[20:],
		},
	}

	// 初期化
	leaf.reset()

	return leaf, nil
}

func (l *Leaf) reset() {
	// prevID, nextIDをInvalidPageIDにセット
	copy(l.header.prevID, to8Bytes(disk.InvalidPageID))
	copy(l.header.nextID, to8Bytes(disk.InvalidPageID))

	// スロット数0、空きスペースは全てのボディ
	l.body.reset()
}

func (l *Leaf) GetPrevID() disk.PageID {
	return toPageID(l.header.prevID)
}

func (l *Leaf) GetNextID() disk.PageID {
	return toPageID(l.header.nextID)
}

func (l *Leaf) GetNumSlots() uint16 {
	return binary.LittleEndian.Uint16(l.body.header.numSlot)
}

func (l *Leaf) GetFreeSpace() uint16 {
	return binary.LittleEndian.Uint16(l.body.header.freeSpace)
}

// ======================================================================

// type BranchHeader struct {
// 	rightID disk.PageID
// }

// type Branch struct {
// 	header BranchHeader
// 	body   Slot
// }

// ======================================================================

type BTree struct {
	metaID disk.PageID // メタデータのページID
}

// 生成される[metaPage]と[rootPage]は、btreeが存在する限り、常に存在する（unpinされない）
func NewBTree(poolManager *pool.PoolManager) (*BTree, error) {
	// メタページ作成
	metaID, err := poolManager.CreatePage()
	if err != nil {
		return nil, err
	}
	metaPage, err := poolManager.FetchPage(metaID)
	if err != nil {
		return nil, err
	}

	// ルートページ作成
	rootID, err := poolManager.CreatePage()
	if err != nil {
		return nil, err
	}
	// ルートページ取得
	rootPage, err := poolManager.FetchPage(rootID)
	if err != nil {
		return nil, err
	}

	// メタページからメタデータ取得
	metaData, err := NewMeta(metaPage)
	if err != nil {
		return nil, err
	}
	// 初期化
	if err = metaData.SetRootID(rootID); err != nil {
		return nil, err
	}

	// ルートページからルートノード取得
	rootNode, err := NewNode(rootPage)
	if err != nil {
		return nil, err
	}
	// 初期化: ルートノードのノードタイプをセット (ルートノードも最初はリーフノード)
	rootNode.SetNodeType(LeafNodeType)

	// ルートノードからリーフノード取得と初期化
	_, err = NewLeaf(rootNode)
	if err != nil {
		return nil, err
	}

	return &BTree{
		metaID: metaID, // メタデータのページIDはここでセットするから、SetMetaID()は不要
	}, nil
}

func (b *BTree) GetMetaID() disk.PageID {
	return b.metaID
}

// BTreeによって確保されているページを全てアンピンし、メタデータのページIDを無効値にする
func (b *BTree) Clear(poolManager *pool.PoolManager) error {
	metaPage, err := poolManager.FetchPage(b.metaID)
	if err != nil {
		return err
	}
	// ここで作成したページとBtree作成時に作ったページをアンピン
	defer metaPage.Unpin()
	defer metaPage.Unpin()

	metaData, err := NewMeta(metaPage)
	if err != nil {
		return err
	}

	rootPage, err := poolManager.FetchPage(metaData.GetRootID())
	if err != nil {
		return err
	}
	// ここで作成したページとBtree作成時に作ったページをアンピン
	defer rootPage.Unpin()
	defer rootPage.Unpin()

	b.metaID = disk.InvalidPageID

	return nil
}
