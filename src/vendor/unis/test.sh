
sed -i ""  's/fw.ErrorLogger.Printf/glog.Errorf/g'   *.go
sed -i ""  's/fw.DebugLogger.Printf/glog.Infof/g'   *.go
sed -i ""  's/fw.InfoLogger.Print/glog.Infof/g'   *.go
sed -i ""  's/fw.WarnLogger.Printf/glog.Warningf/g'   *.go


sed -i ""  's/duxFramework.ErrorLogger.Printf/glog.Errorf/g'   *.go
sed -i ""  's/duxFramework.DebugLogger.Printf/glog.Infof/g'   *.go
sed -i ""  's/duxFramework.InfoLogger.Printf/glog.Infof/g'   *.go
sed -i ""  's/duxFramework.WarnLogger.Printf/glog.Warningf/g'   *.go
