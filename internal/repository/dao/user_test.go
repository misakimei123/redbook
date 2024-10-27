package dao

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGormUserDao_Insert(t *testing.T) {
	now := time.Now().UnixMilli()
	email := sql.NullString{
		String: "123@123.net",
		Valid:  true,
	}
	password := sql.NullString{
		String: "123pass@",
		Valid:  true,
	}
	phone := sql.NullString{
		String: "123",
		Valid:  true,
	}
	wechat_open_id := sql.NullString{
		String: "",
		Valid:  true,
	}
	wechat_union_id := sql.NullString{
		String: "",
		Valid:  true,
	}

	expectedSQL := "INSERT INTO `users` .*"

	testCases := []struct {
		name      string
		sqlmock   func(t *testing.T) *sql.DB
		ctx       context.Context
		user      *User
		wantError error
	}{
		{
			name: "insert success",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				mockRes := sqlmock.NewResult(1, 1)
				expectedArgs := []driver.Value{email, password, now, now, phone, wechat_open_id, wechat_union_id}
				mock.ExpectExec(expectedSQL).
					WithArgs(expectedArgs...).
					WillReturnResult(mockRes)
				return db
			},
			ctx: context.Background(),
			user: &User{
				Email:         email,
				Password:      password,
				Ctime:         now,
				Utime:         now,
				Phone:         phone,
				WechatOpenId:  wechat_open_id,
				WechatUnionId: wechat_union_id,
			},
			wantError: nil,
		},
		{
			name: "duplicate email",
			sqlmock: func(t *testing.T) *sql.DB {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				expectedArgs := []driver.Value{email, password, now, now, phone, wechat_open_id, wechat_union_id}
				mock.ExpectExec(expectedSQL).
					WithArgs(expectedArgs...).
					WillReturnError(&mysqlDriver.MySQLError{Number: 1062})
				return db
			},
			user: &User{
				Email:         email,
				Password:      password,
				Ctime:         now,
				Utime:         now,
				Phone:         phone,
				WechatOpenId:  wechat_open_id,
				WechatUnionId: wechat_union_id,
			},
			ctx:       context.Background(),
			wantError: ErrDuplicateUser,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlDB := tc.sqlmock(t)
			//defer sqlDB.Close()
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      sqlDB,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				SkipDefaultTransaction: true,
				DisableAutomaticPing:   true,
			})
			assert.NoError(t, err)
			dao := NewGormUserDao(db)
			err = dao.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantError, err)
		})
	}
}
