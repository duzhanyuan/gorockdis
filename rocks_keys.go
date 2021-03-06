package main

import (
    "bytes"
    rocks "github.com/tecbot/gorocksdb"
)

func (rh *RocksDBHandler) RedisDel(key []byte, keys ...[]byte) (int, error) {
    if rh.db == nil {
        return 0, ErrRocksIsDead
    }
    if key == nil || len(key) == 0 {
        return 0, ErrWrongArgumentsCount
    }

    keyData := append([][]byte{key}, keys...)
    count := 0
    readOptions := rocks.NewDefaultReadOptions()
    writeOptions := rocks.NewDefaultWriteOptions()
    defer readOptions.Destroy()
    defer writeOptions.Destroy()

    for _, dKey := range keyData {
        _, err := rh.loadRedisObject(readOptions, dKey)
        if err == nil {
            batch := rocks.NewWriteBatch()
            batch.Delete(rh.getTypeKey(dKey))
            batch.Delete(dKey)
            if err := rh.db.Write(writeOptions, batch); err == nil {
                count++
            }
            batch.Destroy()
        }
    }
    return count, nil
}

func (rh *RocksDBHandler) RedisType(key []byte) ([]byte, error) {
    if rh.db == nil {
        return nil, ErrRocksIsDead
    }
    if key == nil || len(key) == 0 {
        return nil, ErrWrongArgumentsCount
    }

    options := rocks.NewDefaultReadOptions()
    defer options.Destroy()

    obj, err := rh.loadRedisObject(options, key)
    if err == nil {
        return []byte(obj.Type), nil
    }
    if err == ErrDoesNotExist {
        return []byte("none"), nil
    }
    return nil, err
}

func (rh *RocksDBHandler) RedisExists(key []byte) (int, error) {
    if rh.db == nil {
        return 0, ErrRocksIsDead
    }
    if key == nil || len(key) == 0 {
        return 0, ErrWrongArgumentsCount
    }
    options := rocks.NewDefaultReadOptions()
    defer options.Destroy()

    if _, err := rh.loadRedisObject(options, key); err == nil {
        return 1, nil
    } else {
        if err == ErrDoesNotExist {
            return 0, nil
        }
        return 0, err
    }
}

// Only support key prefix or all keys, e.g. "KEYS *" or "KEYS test*"
func (rh *RocksDBHandler) RedisKeys(pattern []byte) ([][]byte, error) {
    if rh.db == nil {
        return nil, ErrRocksIsDead
    }
    if pattern == nil || len(pattern) == 0 {
        return nil, ErrWrongArgumentsCount
    }
    if pattern[len(pattern)-1] == '*' {
        pattern = pattern[:len(pattern)-1]
    }

    options := rocks.NewDefaultReadOptions()
    defer options.Destroy()
    options.SetFillCache(false)

    data := make([][]byte, 0)
    it := rh.db.NewIterator(options)
    defer it.Close()
    it.Seek(pattern)
    for ; it.Valid(); it.Next() {
        key := it.Key()
        dKey := rh.copySlice(key, false)
        if bytes.HasPrefix(dKey, kTypeKeyPrefix) {
            continue
        }
        if !bytes.HasPrefix(dKey, pattern) {
            break
        }
        data = append(data, dKey)
    }
    if err := it.Err(); err != nil {
        return nil, err
    }
    return data, nil
}

// This is a dummy stub, since we are using this on redis.
func (rh *RocksDBHandler) RedisExpire(key []byte, timeout int) (int, error) {
    return 1, nil
}
