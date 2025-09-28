package login

/*
// This is for loging in on the frontend with a token
func adminTokenLogin(res http.ResponseWriter, req *http.Request) {
	profile, token, err := helpers.HandleTokenLogin(environment.GetOauth(), res, req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		helpers.ErrorWrapper(res, req, err.Error(), 400)
		return
	}

	adminObj, err := admin.GetByEmail(req.Context(), profile.Email)
	// Query error occured
	if err != nil {
		log.ErrorContext(err, req.Context())
		helpers.ErrorWrapper(res, req, err.Error(), 400)
		return
	}

	// create a session and set its value as the same as the token
	userSession := session.New(tools.ParseStringI(req.Context().Value("ip"))).WithUser(adminObj)
	userSession.Key = helpers.CreateAdminKey(token)
	err = userSession.Save()
	if err != nil {
		log.ErrorContext(err, req.Context())
		helpers.ErrorWrapper(res, req, err.Error(), 400)
		return
	}

	SendSessionCookie(res, environment.GetConfig().Server.AdminSessionKey, userSession.Key)
	successResponse := map[string]interface{}{
		"token": userSession.Key,
	}
	helpers.JSONDataResponseWrapper(res, req, successResponse)
}

func logout(res http.ResponseWriter, req *http.Request) {
	data := make(map[string]any)
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		log.ErrorContext(err, req.Context())
		helpers.ErrorWrapper(res, req, err.Error(), 400)
		return
	}

	if data["token"] != nil {
		userSession := session.Load(tools.ParseStringI(data["token"]))

		if !tools.Empty(userSession) {
			err := userSession.Invalidate()
			if err != nil {
				log.ErrorContext(err, req.Context())
			}
		}
	}

	successResponse := map[string]any{}
	helpers.JSONDataResponseWrapper(res, req, successResponse)
}
*/
