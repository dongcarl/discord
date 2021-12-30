package database

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"

	log "maunium.net/go/maulogger/v2"
	"maunium.net/go/mautrix/id"
)

type User struct {
	db  *Database
	log log.Logger

	MXID id.UserID
	ID   string

	ManagementRoom id.RoomID

	Session *discordgo.Session
}

func (u *User) Scan(row Scannable) *User {
	var token sql.NullString

	err := row.Scan(&u.MXID, &u.ID, &u.ManagementRoom, &token)
	if err != nil {
		if err != sql.ErrNoRows {
			u.log.Errorln("Database scan failed:", err)
		}

		return nil
	}

	if token.Valid {
		session, err := discordgo.New("Bearer " + token.String)
		if err != nil {
			u.log.Errorln("Failed to create discord session:", err)
		} else {
			u.Session = session
		}
	}

	return u
}

func (u *User) sessionNonptr() discordgo.Session {
	if u.Session != nil {
		return *u.Session
	}

	return discordgo.Session{}
}

func (u *User) Insert() {
	session := u.sessionNonptr()

	query := "INSERT INTO user" +
		" (mxid, id, management_room, token)" +
		" VALUES ($1, $2, $3, $4);"

	_, err := u.db.Exec(query, u.MXID, u.ID, u.ManagementRoom,
		session.Identify.Token)

	if err != nil {
		u.log.Warnfln("Failed to insert %s: %v", u.MXID, err)
	}
}

func (u *User) Update() {
	session := u.sessionNonptr()

	query := "UPDATE user SET" +
		" id=$1, management_room=$2, token=$3" +
		" WHERE mxid=$4;"

	_, err := u.db.Exec(query, u.ID, u.ManagementRoom, session.Identify.Token, u.MXID)

	if err != nil {
		u.log.Warnfln("Failed to update %q: %v", u.MXID, err)
	}
}
