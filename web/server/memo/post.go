package memo

import (
	"fmt"
	"git.jasonc.me/main/memo/app/auth"
	"git.jasonc.me/main/memo/app/db"
	"git.jasonc.me/main/memo/app/profile"
	"git.jasonc.me/main/memo/app/res"
	"github.com/cpacia/btcd/chaincfg/chainhash"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/web"
	"net/http"
)

var postRoute = web.Route{
	Pattern:    res.UrlMemoPost + "/" + urlTxHash.UrlPart(),
	Handler: func(r *web.Response) {
		offset := r.Request.GetUrlParameterInt("offset")
		txHashString := r.Request.GetUrlNamedQueryVariable(urlTxHash.Id)
		txHash, err := chainhash.NewHashFromStr(txHashString)
		if err != nil {
			r.Error(jerr.Get("error getting transaction hash", err), http.StatusInternalServerError)
			return
		}
		var pkHash []byte
		if auth.IsLoggedIn(r.Session.CookieId) {
			user, err := auth.GetSessionUser(r.Session.CookieId)
			if err != nil {
				r.Error(jerr.Get("error getting session user", err), http.StatusInternalServerError)
				return
			}
			key, err := db.GetKeyForUser(user.Id)
			if err != nil {
				r.Error(jerr.Get("error getting key for user", err), http.StatusInternalServerError)
				return
			}
			pkHash = key.PkHash
		}
		post, err := profile.GetPostByTxHash(txHash.CloneBytes(), pkHash, uint(offset))
		if err != nil {
			r.Error(jerr.Get("error getting post", err), http.StatusInternalServerError)
			return
		}
		err = profile.AttachLikesToPosts(append(post.Replies, post))
		if err != nil {
			r.Error(jerr.Get("error attaching likes to posts", err), http.StatusInternalServerError)
			return
		}
		r.Helper["Post"] = post
		r.Helper["Title"] = fmt.Sprintf("Memo - Post by %s", post.Name)
		if post.Name == "" {
			r.Helper["Title"] = fmt.Sprintf("Memo - Post by %.6s", post.Memo.GetAddressString())
		}
		r.Helper["Description"] = post.Memo.Message
		res.SetPageAndOffset(r, offset)
		r.RenderTemplate(res.TmplMemoPost)
	},
}
