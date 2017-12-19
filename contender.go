package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// todo: pointers or data as is?
type Contender struct {
	FbId               int `facebook:",required"`
	FbGroupId          int
	Name               string
	TotalPosts         []int
	AvgLikesPerPost    float64 // todo: or float32?
	TotalLikesReceived int
	TotalLikesGiven    int
	PostsUsed          []int
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// SetCreatedAt will set the CreatedAt attribute of a User struct
func (c *Contender) SetCreatedAt(t time.Time) {
	c.CreatedAt = t
}

// SetUpdatedAt will set the UpdatedAt attribute of a User struct
func (c *Contender) SetUpdatedAt(t time.Time) {
	c.UpdatedAt = t
}

// /////////////////
// old methods
// /////////////////

// Sort interface, http://stackoverflow.com/questions/19946992/sorting-a-map-of-structs-golang
type contenderSlice []*Contender

// Len is part of sort.Interface.
func (c contenderSlice) Len() int {
	return len(c)
}

// Swap is part of sort.Interface.
func (c contenderSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Less is part of sort.Interface. Use AvgLikesPerPost as the value to sort by
func (c contenderSlice) Less(i, j int) bool {
	return c[i].AvgLikesPerPost > c[j].AvgLikesPerPost
}

// UpdateContender updates the dependent data fields and updatedAt on the contender's db row
func (c *Contender) UpdateContender(tx *sql.Tx) (int64, error) {
	posts, err := json.Marshal(c.TotalPosts)
	if err != nil {
		return 0, err
	}

	q := `UPDATE contenders SET TotalPosts = ?, TotalLikesReceived = ?, AvgLikesPerPost = ?, TotalLikesGiven = ?, UpdatedAt = CURRENT_TIMESTAMP WHERE Id = ?`
	result, err := tx.Exec(q, posts, c.TotalLikesReceived, c.AvgLikesPerPost, c.TotalLikesGiven, c.FbId)
	if err != nil {
		log.Println(fmt.Sprintf("Failed to update %s's row: %v", c.Name, err))
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// UpdateHDMContenderDependentData creates a map of Contenders to update
//
// For each Post it updates the posting Contender's TotalPosts, TotalLikesReceived,
// and AvgLikesPerPost. Then for each like on the Post it updates those
// Contender's TotalLikesGiven. A transaction then updates each Contender
// in the map to be updated.
func UpdateHDMContenderDependentData() {
	//posts, err := GetHDMPosts(db)
	//if err != nil {
	//	log.Println("Could not get posts from db:", err)
	//	os.Exit(3)
	//}

	// todo: use contender_controller.db
	var db *sql.DB
	var posts []Post

	// Initialize a map of Contenders to be updated
	contenders := make(map[string]Contender)

	// key, value: Post.Id, Post
	for _, p := range posts {
		// log.Println(fmt.Sprintf("Updating %s's post, %v", p.Author, p.Id))

		// Grab contender from db if it has not been updated yet
		var poster *Contender
		if val, ok := contenders[p.Author]; ok {
			poster = &val
		} else {
			//_, err := GetContenderByUsername(db, p.Author)
			//if err != nil {
			//	// if post's author is no longer in the herp, skip it
			//	continue
			//}
		}

		// Update Contender's data fields with Post data
		//poster.TotalPosts = append(poster.TotalPosts, p.Id)
		likesReceived := 0
		//for i := 0; i < len(poster.TotalPosts); i++ {
		//	likesReceived = len(posts[poster.TotalPosts[i]].Likes.Data) + likesReceived
		//}
		poster.TotalLikesReceived = likesReceived
		poster.AvgLikesPerPost = float64(poster.TotalLikesReceived / len(poster.TotalPosts))
		contenders[poster.Name] = *poster

		// For each Post like, give a likes given to the contenders
		for j := 0; j < len(p.Likes.Data); j++ {
			var liker *Contender
			if val, ok := contenders[p.Likes.Data[j].Name]; ok {
				liker = &val
			} else {
				//liker, _ = GetContenderByUsername(db, p.Likes.Data[j].Name)
			}

			// only update likes given for those in the herp
			if liker != nil {
				liker.TotalLikesGiven++
				contenders[liker.Name] = *liker
			}
		}
	}
	log.Println("Finished creating map of Contenders to update")

	// Update every Contender in db that was effected by Posts
	tx, err := db.Begin()
	if err != nil {
		log.Println("Could not begin transaction to update contenders:", err)
		os.Exit(3)
	}
	defer tx.Rollback()

	// key, value: Contender.Id, Contender
	for _, c := range contenders {
		_, err := c.UpdateContender(tx)
		if err != nil {
			log.Println("Could not update contender,", err)
			os.Exit(3)
		}
	}

	if err = tx.Commit(); err != nil {
		log.Println("Failed to COMMIT txn:", err)
	}
	log.Println("Updated Contender dependent data")
}
