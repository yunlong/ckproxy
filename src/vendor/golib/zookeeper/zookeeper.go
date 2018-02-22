package zookeeper

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
)

var _ = fmt.Print

// Create creates the path with value into the zookeeper whether it exists or not
func Create(conn *zk.Conn, path string, data []byte, flags int32, acl []zk.ACL) error {
	path = strings.TrimRight(path, "/")
	nn := strings.Split(path, "/")
	fmt.Printf("%v\n", nn)
	prefix := ""
	for i := 0; i < len(nn); i++ {
		fmt.Printf("======> path=[%v] prefix=[%v] i=%d [%v]\n", path, prefix, i, nn[i])
		if len(nn[i]) == 0 {
			continue
		}
		prefix = prefix + "/" + nn[i]

		exist, stat, err := conn.Exists(prefix)
		fmt.Printf("%v exist=%v err=%v\n", prefix, exist, err)
		if err != nil {
			return fmt.Errorf("check the path [%s] whether existed ERROR:%v", prefix, err.Error())
		}

		if prefix == path {
			if exist {
				conn.Delete(prefix, stat.Version)
			}
			_, err := conn.Create(prefix, data, flags, acl)
			if err != nil {
				return fmt.Errorf("create the path [%s] with node value failed  :%v", prefix, err.Error())
			}
		} else {
			if exist {
				continue
			}

			_, err := conn.Create(prefix, []byte{}, 0, acl)
			if err != nil {
				return fmt.Errorf("create the path [%s] failed :%v", prefix, err.Error())
			} else {
				fmt.Printf("created path [%v] successfully.\n", prefix)
			}
		}
	}
	return nil
}
