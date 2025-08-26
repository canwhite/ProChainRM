// 可以直接调用network包中的方法，只需要在本文件顶部通过import引入network包即可，例如：
// import "sdk-go/network" //包的层级体现出来就ok了
// 然后就可以直接调用network.NewGrpcConnection()、network.NewIdentity()、network.NewSign()等方法。
// 例如：
// conn, err := network.NewGrpcConnection()
// id := network.NewIdentity()
// sign := network.NewSign()