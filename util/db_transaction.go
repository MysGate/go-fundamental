package util

import "github.com/go-xorm/xorm"

type dbTransactionFunc func(sess *xorm.Session) error

func InTransaction(db *xorm.Engine, callback dbTransactionFunc) error {
	var err error
	sess := db.NewSession()
	defer sess.Close()

	if err = sess.Begin(); err != nil {
		return err
	}

	err = callback(sess)
	if err != nil {
		sess.Rollback()
		return err
	} else if err = sess.Commit(); err != nil {
		return err
	}
	return nil
}
