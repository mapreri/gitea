// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package issues

import (
	"testing"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/models/unittest"
	"code.gitea.io/gitea/modules/timeutil"

	"github.com/stretchr/testify/assert"
)

func TestContentHistory(t *testing.T) {
	assert.NoError(t, unittest.PrepareTestDatabase())

	dbCtx := db.DefaultContext
	timeStampNow := timeutil.TimeStampNow()

	_ = SaveIssueContentHistory(dbCtx, 1, 10, 0, timeStampNow, "i-a", true)
	_ = SaveIssueContentHistory(dbCtx, 1, 10, 0, timeStampNow.Add(2), "i-b", false)
	_ = SaveIssueContentHistory(dbCtx, 1, 10, 0, timeStampNow.Add(7), "i-c", false)

	_ = SaveIssueContentHistory(dbCtx, 1, 10, 100, timeStampNow, "c-a", true)
	_ = SaveIssueContentHistory(dbCtx, 1, 10, 100, timeStampNow.Add(5), "c-b", false)
	_ = SaveIssueContentHistory(dbCtx, 1, 10, 100, timeStampNow.Add(20), "c-c", false)
	_ = SaveIssueContentHistory(dbCtx, 1, 10, 100, timeStampNow.Add(50), "c-d", false)
	_ = SaveIssueContentHistory(dbCtx, 1, 10, 100, timeStampNow.Add(51), "c-e", false)

	h1, _ := GetIssueContentHistoryByID(dbCtx, 1)
	assert.EqualValues(t, 1, h1.ID)

	m, _ := QueryIssueContentHistoryEditedCountMap(dbCtx, 10)
	assert.Equal(t, 3, m[0])
	assert.Equal(t, 5, m[100])

	/*
		we can not have this test with real `User` now, because we can not depend on `User` model (circle-import), so there is no `user` table
		when the refactor of models are done, this test will be possible to be run then with a real `User` model.
	*/
	type User struct {
		ID       int64
		Name     string
		FullName string
	}
	_ = db.GetEngine(dbCtx).Sync2(&User{})

	list1, _ := FetchIssueContentHistoryList(dbCtx, 10, 0)
	assert.Len(t, list1, 3)
	list2, _ := FetchIssueContentHistoryList(dbCtx, 10, 100)
	assert.Len(t, list2, 5)

	hasHistory1, _ := HasIssueContentHistory(dbCtx, 10, 0)
	assert.True(t, hasHistory1)
	hasHistory2, _ := HasIssueContentHistory(dbCtx, 10, 1)
	assert.False(t, hasHistory2)

	h6, h6Prev, _ := GetIssueContentHistoryAndPrev(dbCtx, 6)
	assert.EqualValues(t, 6, h6.ID)
	assert.EqualValues(t, 5, h6Prev.ID)

	// soft-delete
	_ = SoftDeleteIssueContentHistory(dbCtx, 5)
	h6, h6Prev, _ = GetIssueContentHistoryAndPrev(dbCtx, 6)
	assert.EqualValues(t, 6, h6.ID)
	assert.EqualValues(t, 4, h6Prev.ID)

	// only keep 3 history revisions for comment_id=100, the first and the last should never be deleted
	keepLimitedContentHistory(dbCtx, 10, 100, 3)
	list1, _ = FetchIssueContentHistoryList(dbCtx, 10, 0)
	assert.Len(t, list1, 3)
	list2, _ = FetchIssueContentHistoryList(dbCtx, 10, 100)
	assert.Len(t, list2, 3)
	assert.EqualValues(t, 8, list2[0].HistoryID)
	assert.EqualValues(t, 7, list2[1].HistoryID)
	assert.EqualValues(t, 4, list2[2].HistoryID)
}
