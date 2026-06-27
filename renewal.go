package main

import (
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hrfee/mediabrowser"
	"github.com/lithammer/shortuuid/v3"
)

func calculateRenewalExpiry(now time.Time, existing *UserExpiry, invite Invite) (time.Time, bool) {
	if !invite.UserExpiry || (invite.UserMonths <= 0 && invite.UserDays <= 0 && invite.UserHours <= 0 && invite.UserMinutes <= 0) {
		return time.Time{}, false
	}

	base := now
	if existing != nil && existing.Expiry.After(now) {
		base = existing.Expiry
	}

	return base.AddDate(0, invite.UserMonths, invite.UserDays).Add(time.Duration(60*invite.UserHours+invite.UserMinutes) * time.Minute), true
}

func normalizeInviteCode(input string) string {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil || parsed.Path == "" || (!strings.Contains(raw, "/") && parsed.Scheme == "") {
		return raw
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] != "invite" {
			continue
		}
		if code, err := url.PathUnescape(parts[i+1]); err == nil {
			return strings.TrimSpace(code)
		}
		return strings.TrimSpace(parts[i+1])
	}

	last := parts[len(parts)-1]
	if code, err := url.PathUnescape(last); err == nil {
		return strings.TrimSpace(code)
	}
	return strings.TrimSpace(last)
}

func (app *appContext) renewalEmailMatches(jellyfinID string, email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}

	stored, ok := app.storage.GetEmailsKey(jellyfinID)
	return ok && stored.Addr != "" && strings.EqualFold(strings.TrimSpace(stored.Addr), email)
}

func (app *appContext) renewAccountWithInvite(gc *gin.Context, code string, user mediabrowser.User, sourceType ActivitySource, source string) (time.Time, bool) {
	code = normalizeInviteCode(code)
	if code == "" || !app.checkInvite(code, false, "") {
		respond(401, "errorInvalidCode", gc)
		return time.Time{}, false
	}

	invite, ok := app.storage.GetInvitesKey(code)
	if !ok {
		respond(401, "errorInvalidCode", gc)
		return time.Time{}, false
	}

	var existing *UserExpiry
	if exp, ok := app.storage.GetUserExpiryKey(user.ID); ok {
		existing = &exp
	}

	expiry, ok := calculateRenewalExpiry(time.Now(), existing, invite)
	if !ok {
		respond(400, "errorInviteNoRenewalDuration", gc)
		return time.Time{}, false
	}

	if user.Policy.IsDisabled {
		if err, changed, activityType := app.SetUserDisabled(user, false); err != nil {
			respond(500, "errorUnknown", gc)
			return time.Time{}, false
		} else if changed {
			app.storage.SetActivityKey(shortuuid.New(), Activity{
				Type:       activityType,
				UserID:     user.ID,
				SourceType: sourceType,
				Source:     source,
				InviteCode: code,
				Time:       time.Now(),
			}, gc, sourceType == ActivityUser)
		}
	}

	app.storage.SetUserExpiryKey(user.ID, UserExpiry{Expiry: expiry})
	app.checkInvite(code, true, user.Name)
	app.InvalidateWebUserCache()

	return expiry, true
}

// @Summary Renews an existing user via invite code
// @Produce json
// @Param renewUserDTO body renewUserDTO true "Renew user request object"
// @Success 200 {object} renewUserResponse
// @Failure 400 {object} stringResponse
// @Failure 401 {object} stringResponse
// @Failure 500 {object} stringResponse
// @Router /user/renew [post]
// @tags Users
func (app *appContext) RenewUserFromInvite(gc *gin.Context) {
	var req renewUserDTO
	gc.BindJSON(&req)

	user, err := app.jf.UserByName(strings.TrimSpace(req.Username), false)
	if err != nil || !app.renewalEmailMatches(user.ID, req.Email) {
		respond(401, "errorRenewalFailed", gc)
		return
	}

	expiry, ok := app.renewAccountWithInvite(gc, req.Code, user, ActivityAnon, "")
	if !ok {
		return
	}

	gc.JSON(200, renewUserResponse{Expiry: expiry.Unix()})
}

// @Summary Renews the logged-in user's account via invite code
// @Produce json
// @Param myRenewDTO body myRenewDTO true "Renew current user request object"
// @Success 200 {object} renewUserResponse
// @Failure 400 {object} stringResponse
// @Failure 401 {object} stringResponse
// @Failure 500 {object} stringResponse
// @Router /my/renew [post]
// @Security Bearer
// @tags User Page
func (app *appContext) RenewMyAccount(gc *gin.Context) {
	var req myRenewDTO
	gc.BindJSON(&req)

	user, err := app.jf.UserByID(gc.GetString("jfId"), false)
	if err != nil {
		respond(401, "errorRenewalFailed", gc)
		return
	}

	expiry, ok := app.renewAccountWithInvite(gc, req.Code, user, ActivityUser, user.ID)
	if !ok {
		return
	}

	gc.JSON(200, renewUserResponse{Expiry: expiry.Unix()})
}
