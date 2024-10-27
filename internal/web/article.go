package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	intrv1 "github.com/misakimei123/redbook/api/proto/gen/intr/v1"
	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/service"
	"github.com/misakimei123/redbook/internal/web/jwt"
	"github.com/misakimei123/redbook/internal/web/result"
	"github.com/misakimei123/redbook/pkg/logger"
	"golang.org/x/sync/errgroup"
)

type ArticleHandler struct {
	svc            service.ArticleService
	interactiveSvc intrv1.InteractiveServiceClient //service2.InteractiveService
	rankingSvc     service.RankingService
	l              logger.LoggerV1
	bizStr         string
}

func NewArticleHandler(articleService service.ArticleService, interactive intrv1.InteractiveServiceClient, log logger.LoggerV1, rankingService service.RankingService) *ArticleHandler {
	return &ArticleHandler{
		svc:            articleService,
		interactiveSvc: interactive,
		bizStr:         "article",
		rankingSvc:     rankingService,
		l:              log,
	}
}

func (a *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	group := server.Group("/articles")
	group.POST("edit", a.Edit)
	group.POST("publish", a.Publish)
	group.POST("withdraw", a.Withdraw)
	group.POST("list", a.List)
	group.GET("detail/:id", a.Detail)
	group.GET("rank", a.Rank)

	pub := server.Group("/pub")
	pub.GET("/:id", a.PubDetail)
	pub.POST("/like", a.Like)
	pub.POST("/collect", a.Collect)
}

func (a *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	id, err := a.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author:  domain.Author{Id: uc.Uid},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	ctx.JSON(http.StatusOK, result.Result{
		Data: id,
	})
}

func (a *ArticleHandler) Publish(ctx *gin.Context) {
	type Req struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	id, err := a.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author:  domain.Author{Id: uc.Uid},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	ctx.JSON(http.StatusOK, result.Result{
		Data: id,
	})
}

func (a *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	userClaims := ctx.MustGet("user").(jwt.UserClaims)
	err = a.svc.Withdraw(ctx, req.Id, userClaims.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	ctx.JSON(http.StatusOK, result.RetSuccess)
}

func (a *ArticleHandler) Rank(ctx *gin.Context) {
	articles, err := a.rankingSvc.GetTopN(ctx.Request.Context())
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, result.Result{Data: slice.Map[domain.Article, ArticleVo](articles, func(idx int, src domain.Article) ArticleVo {
		return ArticleVo{
			Id:       src.Id,
			Title:    src.Title,
			Abstract: src.Abstract(),
			Status:   src.Status.ToInt(),
			Ctime:    src.Ctime.Format(time.DateTime),
			Utime:    src.Utime.Format(time.DateTime),
		}
	})})
}

func (a *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
	}
	userClaims := ctx.MustGet("user").(jwt.UserClaims)
	articles, err := a.svc.GetByAuthor(ctx, userClaims.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	ctx.JSON(http.StatusOK, result.Result{Data: slice.Map[domain.Article, ArticleVo](articles, func(idx int, src domain.Article) ArticleVo {
		return ArticleVo{
			Id:       src.Id,
			Title:    src.Title,
			Abstract: src.Abstract(),
			Status:   src.Status.ToInt(),
			Ctime:    src.Ctime.Format(time.DateTime),
			Utime:    src.Utime.Format(time.DateTime),
		}
	})})
}

func (a *ArticleHandler) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetIllegalNumberFormat)
		return
	}
	userClaims := ctx.MustGet("user").(jwt.UserClaims)
	article, err := a.svc.GetById(ctx.Request.Context(), userClaims.Uid, id)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	ctx.JSON(http.StatusOK, result.Result{Data: ArticleVo{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
		Status:   article.Status.ToInt(),
		Ctime:    article.Ctime.Format(time.DateTime),
		Utime:    article.Utime.Format(time.DateTime),
	}})
}

func (a *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetIllegalNumberFormat)
		return
	}

	var (
		eg          errgroup.Group
		article     domain.Article
		interactive *intrv1.Interactive
	)

	userClaims := ctx.MustGet("user").(jwt.UserClaims)

	eg.Go(func() error {
		var er error
		article, er = a.svc.GetPubById(ctx, id, userClaims.Uid)
		return er
	})

	eg.Go(func() error {
		var er error
		resp, er := a.interactiveSvc.Get(ctx, &intrv1.GetRequest{
			BizStr: a.bizStr,
			BizId:  id,
			Uid:    userClaims.Uid,
		}) //a.bizStr, id, userClaims.Uid)
		if er != nil {
			return er
		}
		interactive = resp.Interactive
		return nil
	})

	err = eg.Wait()
	if err != nil {
		//	TODO: record log
		a.l.Error("get pub fail",
			logger.String("biz", a.bizStr),
			logger.Int64("id", article.Id),
			logger.Error(err))
		ctx.JSON(http.StatusOK, result.Result{Data: ArticleVo{
			Id:         article.Id,
			Title:      article.Title,
			Content:    article.Content,
			AuthorId:   article.Author.Id,
			AuthorName: article.Author.Name,
			Status:     article.Status.ToInt(),
			Ctime:      article.Ctime.Format(time.DateTime),
			Utime:      article.Utime.Format(time.DateTime),
			ReadCnt:    0,
			LikeCnt:    0,
			CollectCnt: 0,
			Liked:      false,
			Collected:  false,
		}})
		return
	}

	//go func() {
	//	er := a.interactiveSvc.IncrReadCnt(ctx, a.bizStr, article.Id)
	//	if er != nil {
	//		a.l.Error("incr read cnt fail",
	//			logger.String("biz", a.bizStr),
	//			logger.Int64("id", article.Id),
	//			logger.Error(err))
	//	}
	//}()

	ctx.JSON(http.StatusOK, result.Result{Data: ArticleVo{
		Id:         article.Id,
		Title:      article.Title,
		Content:    article.Content,
		AuthorId:   article.Author.Id,
		AuthorName: article.Author.Name,
		Status:     article.Status.ToInt(),
		Ctime:      article.Ctime.Format(time.DateTime),
		Utime:      article.Utime.Format(time.DateTime),
		ReadCnt:    interactive.ReadCnt,
		LikeCnt:    interactive.LikeCnt,
		CollectCnt: interactive.CollectCnt,
		Liked:      interactive.Liked,
		Collected:  interactive.Collected,
	}})
}

func (a *ArticleHandler) Like(ctx *gin.Context) {
	type Req struct {
		Id   int64 `json:"id,omitempty"`
		Like bool  `json:"like,omitempty"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	userClaims := ctx.MustGet("user").(jwt.UserClaims)
	_, err = a.interactiveSvc.Like(ctx.Request.Context(), &intrv1.LikeRequest{
		BizStr: a.bizStr,
		BizId:  req.Id,
		Uid:    userClaims.Uid,
		Like:   req.Like,
	}) //req.Like, a.bizStr, req.Id, userClaims.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	ctx.JSON(http.StatusOK, result.RetSuccess)
}

func (a *ArticleHandler) Collect(ctx *gin.Context) {
	type Req struct {
		Id  int64 `json:"id,omitempty"`
		Cid int64 `json:"cid,omitempty"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	userClaims := ctx.MustGet("user").(jwt.UserClaims)
	_, err = a.interactiveSvc.Collect(ctx, &intrv1.CollectRequest{
		BizStr: a.bizStr,
		BizId:  req.Id,
		Uid:    userClaims.Uid,
		Cid:    req.Cid,
	}) //a.bizStr, req.Id, req.Cid, userClaims.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, result.RetSystemError)
		return
	}
	ctx.JSON(http.StatusOK, result.RetSuccess)
}
