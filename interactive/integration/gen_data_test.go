package integration

import (
	"math/rand"
	"testing"
	"time"

	"github.com/misakimei123/redbook/interactive/integration/startup"
	"github.com/misakimei123/redbook/interactive/repository/dao"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGenData(t *testing.T) {
	db := startup.InitDB()
	const batchSize = 100
	for i := 0; i < 10; i++ {
		data := make([]dao.Interactive, 0, batchSize)
		now := time.Now().UnixMilli()
		for j := 0; j < batchSize; j++ {
			data = append(data, dao.Interactive{
				Id:         0,
				BizId:      int64(i*batchSize + j + 1),
				BizStr:     "test",
				ReadCnt:    rand.Int63(),
				LikeCnt:    rand.Int63(),
				CollectCnt: rand.Int63(),
				Utime:      now,
				Ctime:      now,
			})
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			err := tx.Create(data).Error
			assert.NoError(t, err)
			return err
		})
		assert.NoError(t, err)
	}
}
