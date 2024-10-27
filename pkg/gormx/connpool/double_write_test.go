package connpool

import (
	"testing"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DoubleWriteTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (s *DoubleWriteTestSuite) SetupSuite() {
	t := s.T()
	src, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:3308)/redbook"))
	assert.NoError(t, err)
	err = src.AutoMigrate(&Interactive{})
	assert.NoError(t, err)
	dst, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:3308)/redbook_intr"))
	assert.NoError(t, err)
	err = dst.AutoMigrate(&Interactive{})
	assert.NoError(t, err)
	doubleWrite, err := gorm.Open(mysql.New(mysql.Config{Conn: &DoubleWritePool{
		src:     src.ConnPool,
		dst:     dst.ConnPool,
		pattern: atomicx.NewValueOf(PatternSrcFirst),
		l:       logger.NewNopLogger(),
	}}))
	assert.NoError(t, err)
	s.db = doubleWrite
}

func (s *DoubleWriteTestSuite) TearDownTest() {

}

func (s *DoubleWriteTestSuite) TestDoubleWriteTest() {
	t := s.T()
	err := s.db.Create(&Interactive{
		BizId:  10086,
		BizStr: "test",
	}).Error
	assert.NoError(t, err)
}

func (s *DoubleWriteTestSuite) TestDoubleWriteTransaction() {
	t := s.T()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&Interactive{
			BizId:  10087,
			BizStr: "test",
		}).Error
	})
	assert.NoError(t, err)
}

func TestDoubleWrite(t *testing.T) {
	suite.Run(t, new(DoubleWriteTestSuite))
}

type Interactive struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_type_id"`
	BizStr     string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	ReadCnt    int64
	CollectCnt int64
	LikeCnt    int64
	Ctime      int64
	Utime      int64
}
