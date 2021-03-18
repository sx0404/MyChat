package Formation

type RoleInfo struct {
	UserID			uint64
	UserName		string
	Password  		string
}

type RoleFriendInfo struct {
	UserID			uint64
	FrinedList		[]RoleFriend
}

type RoleFriend struct {
	UserID			uint64
	UserName		string
}

type RoleMoney struct {
	UserID			uint64
	Coin			int64
	Gold			int64
	Sliver			int64
}

type OfflineChat struct {
	SendID			uint64		//需要发送的对方ID
	FromID			uint64 		//发送方的ID
	FromName		string		//发送方的名字
	Content			string		//内容
}