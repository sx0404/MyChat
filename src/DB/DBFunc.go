package db

import (
	"fmt"
	"test/src/Formation"
)

func GetUser(userName string) Formation.RoleInfo {
	Instance := GetDBInstance()
	var roleInfo Formation.RoleInfo
	err := Instance.db.QueryRow("SELECT * FROM g_user WHERE userName = ?",userName).Scan(&roleInfo.UserID, &roleInfo.UserName, &roleInfo.Password)
	if err != nil{
		fmt.Println("GetUser wrong",err)
	}
	return roleInfo
}

func GetUserByUserID(userID uint64) Formation.RoleInfo {
	Instance := GetDBInstance()
	var roleInfo Formation.RoleInfo
	err := Instance.db.QueryRow("SELECT * FROM g_user WHERE userID = ?",userID).Scan(
		&roleInfo.UserID, &roleInfo.UserName, &roleInfo.Password)
	if err != nil{
		fmt.Println("GetUserByUserID wrong")
	}
	return roleInfo
}

func GetUserIDByUserName(userName string) uint64 {
	Instance := GetDBInstance()
	var userID uint64
	err := Instance.db.QueryRow("SELECT userID FROM g_user WHERE userName = ?",userName).Scan(
		&userID)
	if err != nil {
		fmt.Println("GetUserIDByUserName ",err)
	}
	return userID
}

func GetRoleMoney(userID uint64) Formation.RoleMoney {
	Instance := GetDBInstance()
	var roleMoney Formation.RoleMoney
	err := Instance.db.QueryRow("SELECT  * FROM g_money WHERE userID = ?",userID).Scan(
		&roleMoney.UserID, &roleMoney.Coin, &roleMoney.Gold, &roleMoney.Sliver)
	if err != nil {
		fmt.Println("GetRoleMoney wrong")
	}
	return roleMoney
}

func GetUserName(userID uint64) string {
	Instance := GetDBInstance()
	var userName string
	err := Instance.db.QueryRow("SELECT  * FROM g_user WHERE userID = ?",userID).Scan(
		&userName)
	if err != nil {
		fmt.Println("GetUserName wrong",err)
	}
	return userName
}

func GetRoleFriendInfo(userID uint64) Formation.RoleFriendInfo {
	Instance := GetDBInstance()
	rows ,err := Instance.db.Query("SELECT friendID FROM g_friend WHERE userID = ?",userID)
	var friends []Formation.RoleFriend
	var friendsInfo Formation.RoleFriendInfo
	if err != nil{
		fmt.Println("GetRoleFriendInfo error",err)
		return friendsInfo
	}
	//循环读取结果
	for rows.Next() {
		friend := Formation.RoleFriend{}
		//将每一行的结果都赋值到一个user对象中
		err := rows.Scan(&friend.UserID)
		if err != nil {
			fmt.Println("rows fail")
		}
		friend.UserName = GetUserName(friend.UserID)

		friends = append(friends, friend)
	}
	friendsInfo.UserID = userID
	friendsInfo.FrinedList = friends
	return friendsInfo
}

func InserRoleInfo(info Formation.RoleInfo) {
	instance := GetDBInstance()
	instance.Insert("g_user",[]string{"userID","userName","passWord"},info.UserID,info.UserName,info.Password)
}

func InsertChat(info Formation.OfflineChat) {
	instance := GetDBInstance()
	instance.Insert("g_offline_chat",[]string{"sendID","fromID","fromName","content"},
	info.SendID,info.FromID,info.FromName,info.Content)
}

