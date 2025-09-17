package login

/*
// Does auth0 purely server side login
func authLogout(res http.ResponseWriter, req *http.Request) {
	auth0.HandleAuth0Logout(environment.GetAuth0(), res, req)
}

// Does auth0 purely server side login
func authLogin(res http.ResponseWriter, req *http.Request) {
	auth0.HandleAuth0Login(environment.GetAuth0(), res, req)
}

// Does auth0 purely server side login
func authCallback(res http.ResponseWriter, req *http.Request) {
	profile, err := auth0.HandleAuth0Callback(environment.GetAuth0(), res, req)
	if err != nil {
		log.ErrorRequest(err,res)
		helpers.ErrorWrapper(res, req, err.Error(), 400)
		return
	}

	userObj, err := user.GetByEmail(req.Context(), profile.Email)
	// Query error occured
	if err != nil {
		log.ErrorRequest(err,res)
		helpers.ErrorWrapper(res, req, err.Error(), 400)
		return
	}

	// User not found - invalid login detected
	if tools.Empty(userObj) {
		userObj := user.New()
		userObj.Set("email", profile.Email)
		userObj.Set("name", profile.Name)
		err := userObj.Save(nil)
		if err != nil {
			log.ErrorRequest(err,res)
			helpers.ErrorWrapper(res, req, err.Error(), 400)
			return
		}
	}

	userSession := session.New(userObj, tools.ParseStringI(req.Context().Value("ip")))
	err = userSession.SaveWithExpiration(int64(profile.Exp))
	if err != nil {
		log.ErrorRequest(err,res)
		helpers.ErrorWrapper(res, req, err.Error(), 400)
		return
	}

	SendSessionCookie(res, userSession.Key)
	appUrl := senv.GetConfig().Server.AppURL
	http.Redirect(res, req, appUrl, http.StatusTemporaryRedirect)

}

// This is for loging in on the frontend with a token
func tokenLogin(res http.ResponseWriter, req *http.Request) {
	profile, token, err := auth0.HandleTokenLogin(environment.GetAuth0(), res, req)
	if err != nil {
		log.ErrorRequest(err,res)
		helpers.ErrorWrapper(res, req, err.Error(), 400)
		return
	}

	userObj, err := user.GetByEmail(req.Context(), profile.Email)
	// Query error occured
	if err != nil {
		log.ErrorRequest(err,res)
		helpers.ErrorWrapper(res, req, err.Error(), 400)
		return
	}

	// User not found create them an account
	if tools.Empty(userObj) {
		userObj := user.New()
		userObj.Set("email", profile.Email)
		userObj.Set("name", profile.Name)
		userObj.Set("picture", profile.Picture)
		err := userObj.Save(nil)
		if err != nil {
			log.ErrorRequest(err,res)
			helpers.ErrorWrapper(res, req, err.Error(), 400)
			return
		}
	}

	// create a session and set its value as the same as the token
	userSession := session.New(userObj, tools.ParseStringI(req.Context().Value("ip")))
	userSession.Key = token
	err = userSession.Save()
	if err != nil {
		log.ErrorRequest(err,res)
		helpers.ErrorWrapper(res, req, err.Error(), 400)
		return
	}

	SendSessionCookie(res, userSession.Key)
	successResponse := map[string]interface{}{
		"token":   userSession.Key,
		"user_id": userObj.ID(),
	}
	json.NewEncoder(res).Encode(successResponse)

}
*/
