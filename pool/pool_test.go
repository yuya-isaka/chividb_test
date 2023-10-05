package pool_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuya-isaka/chibidb/disk"
	"github.com/yuya-isaka/chibidb/pool"
)

func createPage(poolManager *pool.PoolManager, bytes []byte) (disk.PageID, error) {
	// ページ作成
	pageID, err := poolManager.CreatePage()
	if err != nil {
		return disk.InvalidID, err
	}

	// ページデータ書き込み
	fetchPage, err := poolManager.FetchPage(pageID)
	if err != nil {
		return disk.InvalidID, err
	}
	fetchPage.SetData(bytes)
	fetchPage.SetUpdate(true)
	defer fetchPage.SubPin()

	return pageID, nil
}

func TestPool(t *testing.T) {
	// 準備
	assert := assert.New(t)

	// テストデータ準備
	helloBytes := make([]byte, disk.PageSize)
	copy(helloBytes, "Hello")
	worldBytes := make([]byte, disk.PageSize)
	copy(worldBytes, "World")

	// ------------------------------------------------------------------

	t.Run("Simple Pool 3", func(t *testing.T) {
		// テストファイル準備
		testFile := "testfile"
		fileManager, err := disk.NewFileManager(testFile)
		assert.NoError(err)
		defer os.Remove(testFile)

		// プール準備
		poolTest := pool.NewPool(3)
		poolManager := pool.NewPoolManager(fileManager, poolTest)
		defer poolManager.Close()

		// ------------------------------------------------------------------

		// create (hello)
		helloID, err := createPage(poolManager, helloBytes)
		assert.NoError(err)

		// fetch (hello)
		fetchPage, err := poolManager.FetchPage(helloID)
		assert.NoError(err)

		// テスト
		assert.Equal(disk.PageID(0), helloID)
		assert.Equal(helloBytes, fetchPage.GetData())
	})

	t.Run("Complex Pool 3", func(t *testing.T) {
		// テストファイル準備
		testFile := "testfile"
		fileManager, err := disk.NewFileManager(testFile)
		assert.NoError(err)
		defer os.Remove(testFile)

		// プール準備
		poolTest := pool.NewPool(3)
		poolManager := pool.NewPoolManager(fileManager, poolTest)
		defer poolManager.Close()

		// ------------------------------------------------------------------

		// create (hello)
		helloID, err := createPage(poolManager, helloBytes)
		assert.NoError(err)

		// fetch (hello)
		fetchPage, err := poolManager.FetchPage(helloID)
		assert.NoError(err)

		// テスト (hello)
		assert.Equal(disk.PageID(0), helloID)
		assert.Equal(helloBytes, fetchPage.GetData())

		// ------------------------------------------------------------------

		// create (world)
		worldID, err := createPage(poolManager, worldBytes)
		assert.NoError(err)

		// ------------------------------------------------------------------

		// fetch (hello)
		fetchPage, err = poolManager.FetchPage(helloID)
		assert.NoError(err)

		// テスト (hello)
		assert.Equal(disk.PageID(0), helloID)
		assert.Equal(helloBytes, fetchPage.GetData())

		// ------------------------------------------------------------------

		// fetch (world)
		fetchPage, err = poolManager.FetchPage(worldID)
		assert.NoError(err)

		// テスト (world)
		assert.Equal(disk.PageID(1), worldID)
		assert.Equal(worldBytes, fetchPage.GetData())
	})

	t.Run("Pool 1", func(t *testing.T) {
		// テストファイル準備
		testFile := "testfile"
		fileManager, err := disk.NewFileManager(testFile)
		assert.NoError(err)
		defer os.Remove(testFile)

		// プール準備
		poolTest := pool.NewPool(1)
		poolManager := pool.NewPoolManager(fileManager, poolTest)
		defer poolManager.Close()

		// ------------------------------------------------------------------

		// create (hello)
		helloID, err := createPage(poolManager, helloBytes)
		assert.NoError(err)

		// fetch (hello)
		fetchPage, err := poolManager.FetchPage(helloID)
		assert.NoError(err)

		// テスト (hello)
		assert.Equal(disk.PageID(0), helloID)
		assert.Equal(helloBytes, fetchPage.GetData())

		// ------------------------------------------------------------------

		// Error test
		// プールのサイズは１で、fetchPageがまだ持っているので、エラーになる
		_, err = poolManager.CreatePage()
		assert.Error(err)

		// 参照カウンタを減らすことで、新しいページが作れるようになる
		// helloPageとfetchPageは同じページを参照しており、そのページのカウントを２回下げることで-1になる
		fetchPage.SubPin()
		assert.Equal(pool.Pin(-1), fetchPage.GetPinCount())

		// ------------------------------------------------------------------

		// create (world)
		worldID, err := createPage(poolManager, worldBytes)
		assert.NoError(err)

		// fetch (world)
		fetchPage, err = poolManager.FetchPage(worldID)
		assert.NoError(err)

		// テスト (world)
		assert.Equal(disk.PageID(1), worldID)
		assert.Equal(worldBytes, fetchPage.GetData())

		// ------------------------------------------------------------------

		// Error test
		_, err = poolManager.CreatePage()
		assert.Error(err)

		fetchPage.SubPin()
		assert.Equal(pool.NoReferencePin, fetchPage.GetPinCount())

		// ------------------------------------------------------------------

		// helloIDはコピーされているので０のままのはず
		assert.Equal(disk.PageID(0), helloID)

		// helloが格納されているpageIDは変わらない
		fetchPage, err = poolManager.FetchPage(helloID)
		assert.NoError(err)

		// テスト (hello)
		assert.Equal(helloBytes, fetchPage.GetData())
	})

	t.Run("Pool 2", func(t *testing.T) {
		// テストファイル準備
		testFile := "testfile"
		fileManager, err := disk.NewFileManager(testFile)
		assert.NoError(err)
		defer os.Remove(testFile)

		// プール準備
		poolTest := pool.NewPool(2)
		poolManager := pool.NewPoolManager(fileManager, poolTest)
		defer poolManager.Close()

		// ------------------------------------------------------------------

		// create (hello)
		helloID, err := createPage(poolManager, helloBytes)
		assert.NoError(err)

		// fetch (hello)
		fetchPage, err := poolManager.FetchPage(helloID)
		assert.NoError(err)

		// テスト (hello)
		assert.Equal(disk.PageID(0), helloID)
		assert.Equal(helloBytes, fetchPage.GetData())

		// ------------------------------------------------------------------

		// create (world)
		worldID, err := createPage(poolManager, worldBytes)
		assert.NoError(err)

		// fetch (world)
		fetchPage, err = poolManager.FetchPage(worldID)
		assert.NoError(err)

		// テスト (world)
		assert.Equal(disk.PageID(1), worldID)
		assert.Equal(worldBytes, fetchPage.GetData())

		// ------------------------------------------------------------------

		// fetch (hello)
		fetchPage, err = poolManager.FetchPage(helloID)
		assert.NoError(err)

		// テスト (hello)
		assert.Equal(helloBytes, fetchPage.GetData())
	})
}