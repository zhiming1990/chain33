package db

import (
	"bytes"

	log "github.com/inconshreveable/log15"
)

type ListHelper struct {
	db IteratorDB
}

var listlog = log.New("module", "db.ListHelper")

func NewListHelper(db IteratorDB) *ListHelper {
	return &ListHelper{db}
}

func (db *ListHelper) PrefixScan(prefix []byte) (values [][]byte) {
	it := db.db.Iterator(prefix, false)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {
		value := it.ValueCopy()
		if it.Error() != nil {
			listlog.Error("PrefixScan it.Value()", "error", it.Error())
			values = nil
			return
		}
		//blog.Debug("PrefixScan", "key", string(item.Key()), "value", string(value))
		values = append(values, value)
	}
	return
}

const (
	ListDESC = int32(0)
	ListASC  = int32(1)
	ListSeek = int32(2)
)

func (db *ListHelper) List(prefix, key []byte, count, direction int32) (values [][]byte) {
	if len(key) == 0 {
		if direction == ListASC {
			return db.IteratorScanFromFirst(prefix, count)
		} else {
			return db.IteratorScanFromLast(prefix, count)
		}
	}
	if count == 1 && direction == ListSeek {
		it := db.db.Iterator(prefix, true)
		defer it.Close()
		it.Seek(key)
		//判断是否相等
		if !bytes.Equal(key, it.Key()) {
			it.Next()
			if !it.Valid() {
				return nil
			}
		}
		return [][]byte{cloneByte(it.Key()), cloneByte(it.Value())}
	}
	return db.IteratorScan(prefix, key, count, direction)
}

func (db *ListHelper) IteratorScan(prefix []byte, key []byte, count int32, direction int32) (values [][]byte) {
	var reserse = false
	if direction == 0 {
		reserse = true
	}
	it := db.db.Iterator(prefix, reserse)
	defer it.Close()

	var i int32
	it.Seek(key)
	if !it.Valid() {
		listlog.Error("PrefixScan it.Value()", "error", it.Error())
		values = nil
		return
	}
	for it.Next(); it.Valid(); it.Next() {
		value := it.ValueCopy()
		if it.Error() != nil {
			listlog.Error("PrefixScan it.Value()", "error", it.Error())
			values = nil
			return
		}
		// blog.Debug("PrefixScan", "key", string(item.Key()), "value", value)
		values = append(values, value)
		i++
		if i == count {
			break
		}
	}
	return
}

func (db *ListHelper) IteratorScanFromFirst(prefix []byte, count int32) (values [][]byte) {
	it := db.db.Iterator(prefix, false)
	defer it.Close()

	var i int32
	for it.Rewind(); it.Valid(); it.Next() {
		value := it.ValueCopy()
		if it.Error() != nil {
			listlog.Error("PrefixScan it.Value()", "error", it.Error())
			values = nil
			return
		}
		//listlog.Debug("PrefixScan", "key", string(it.Key()), "value", value)
		values = append(values, value)
		i++
		if i == count {
			break
		}
	}
	return
}

func (db *ListHelper) IteratorScanFromLast(prefix []byte, count int32) (values [][]byte) {
	it := db.db.Iterator(prefix, true)
	defer it.Close()

	var i int32
	for it.Rewind(); it.Valid(); it.Next() {
		value := it.ValueCopy()
		if it.Error() != nil {
			listlog.Error("PrefixScan it.Value()", "error", it.Error())
			values = nil
			return
		}
		// blog.Debug("PrefixScan", "key", string(item.Key()), "value", value)
		values = append(values, value)
		i++
		if i == count {
			break
		}
	}
	return
}

func (db *ListHelper) PrefixCount(prefix []byte) (count int64) {
	it := db.db.Iterator(prefix, true)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {
		if it.Error() != nil {
			listlog.Error("PrefixCount it.Value()", "error", it.Error())
			count = 0
			return
		}

		count++
	}
	return
}

//for test
func (db *ListHelper) ListExcept(expectPrefix []byte, exceptPrefix []byte, count int32, direction int32) (values [][]byte) {
	var reserse = false
	if direction == 0 {
		reserse = true
	}
	it := db.db.Iterator(expectPrefix, reserse)
	defer it.Close()

	var i int32
	for it.Rewind(); it.Valid(); it.Next() {
		value := it.ValueCopy()
		if it.Error() != nil {
			listlog.Error("PrefixScan it.Value()", "error", it.Error())
			values = nil
			return
		}

		if bytes.Compare(it.Key(), exceptPrefix) >= 0 {
			//fmt.Println(string(it.Key()), "bigger than ", string(exceptPrefix))
			continue
		} else {
			//fmt.Println(string(it.Key()), "less than ", string(exceptPrefix))
		}
		// blog.Debug("PrefixScan", "key", string(item.Key()), "value", value)
		values = append(values, value)
		i++
		if i == count {
			break
		}
	}
	return
}

func (db *ListHelper) ListExceptAndExcute(expectPrefix []byte, exceptPrefix []byte, count int32, direction int32, fn func(key, value []byte) bool) {
	var reserse = false
	if direction == 0 {
		reserse = true
	}

	//针对expectPrefix为".-mvcc-.l.mavl-coins-bty-06htvcBNSEA7fZhAdLJphDwQRQJaHpy001"的情况，需要做特殊处理:
	//先截取最后一个'-'及之前的部分作为前缀扫描，再判断每一个元素的key值是否大于等于expectPrefix，如果是，则进行处理；否则，记录是之前已经处理过的，直接跳过。
	var it Iterator
	startKeyFlag := false
	lastIndexOfDelimiter := bytes.LastIndex(expectPrefix, []byte("-"))
	if lastIndexOfDelimiter+1 < len(expectPrefix) {
		it = db.db.Iterator(expectPrefix[0:lastIndexOfDelimiter+1], reserse)
		startKeyFlag = true
	} else {
		it = db.db.Iterator(expectPrefix, reserse)
	}

	defer it.Close()

	var i int32

	for it.Rewind(); it.Valid(); it.Next() {
		value := it.ValueCopy()
		if it.Error() != nil {
			listlog.Error("PrefixScan it.Value()", "error", it.Error())
			return
		}

		if bytes.Compare(it.Key(), exceptPrefix) >= 0 {
			//fmt.Println(string(it.Key()), "bigger than ", string(exceptPrefix), " ,so exclude it.")
			continue
		} else {
			//fmt.Println(string(it.Key()), "less than ", string(exceptPrefix))
		}

		if startKeyFlag && bytes.Compare(it.Key(), expectPrefix) < 0 {
			//fmt.Println(string(it.Key()), "less than ", string(expectPrefix), " ,so exclude it.")
			continue
		} else {
			//fmt.Println(string(it.Key()), "not less than ", string(expectPrefix))
		}
		key := make([]byte, len(it.Key()))
		copy(key, it.Key())
		if fn(key, value) {
			break
		}
		i++
		if i == count {
			break
		}
	}
}
