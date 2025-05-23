package messages

const (
	GeneralSuccess      = "عملیات با موفقیت انجام شد"
	GeneralFailed       = "عملیات با شکست مواجه شد"
	GeneralAccessDenied = "شما مجاز به انجام این کار نیستید"
	GeneralUnauthorized = "برای دسترسی به این بخش لطفا وارد شوید"
	GeneralBadRequest   = "اطلاعات وارد شده رو مجددا بررسی کنید"
	GeneralNotFound     = "موردی یافت نشد"
	GeneralNeedsPremium = "برای دسترسی به این بخش لطفا اکانت حرفه‌ای تهیه کنید"

	OTPInvalid  = "کد وارد شده صحیح نیست"
	OTPTryAgain = "لطفا بعد از گذشت %d ثانیه مجددا تلاش کنید"

	InvalidRefreshToken = "لطفا مجددا وارد شوید"

	StoryNotFound          = "داستانی یافت نشد"
	StoryCreated           = "داستان با موفقیت ثبت شد"
	StoryNotCreated        = "ثبت داستان موفقیت آمیز نبود"
	StoryEdited            = "داستان با موفقیت ویرایش شد"
	StoryDeleted           = "داستان با موفقیت حذف شد"
	StoryAlreadyLiked      = "این داستان رو قبلا لایک کردید"
	StoryAlreadyBookmarked = "این داستان رو قبلا ذخیره کردید"
	StoryCharLimit         = "داستان باید حداقل ۲۵ و حداکثر ۲۵۰ حرف باشد"
	StoryLiked             = "از این داستان خوشت اومد"
	StoryDisliked          = "با این داستان حال نکردی"
	StoryShareLimit        = "قبلا این داستان رو به اشتراک گذاشتی"
	// StoryMinCharLimit = "داستان باید حداقل شامل ۲۵ حرف باشد"
	// StoryMaxCharLimit = "داستان می‌تواند نهایتا شامل ۲۵۶ حرف باشد"

	CommentNotFound     = "نقدی یافت نشد"
	CommentEdited       = "نقد با موفقیت ویرایش شد"
	CommentLiked        = "از این نقد خوشت اومد"
	CommentDisliked     = "با این نقد حال نکردی"
	CommentAlreadyLiked = "این نقد رو قبلا لایک کردید"
	CommentDeleted      = "نقد با موفقیت حذف شد"
	CommentCharLimit    = "نقد باید حداقل ۵ و حداکثر ۲۵۰ حرف باشد"

	UserNotFound        = "کاربری با این شناسه یافت نشد"
	UserEdited          = "اطلاعات شما با موفقیت ویرایش شد"
	UserForbiddenName   = "لطفا فقط از کلمات فارسی استفاده کنید"
	UserAlreadyFollowed = "این کاربر را قبلا دنبال کرده‌اید"
	UserFollowSelf      = "نمی‌تونید خودتون رو دنبال کنید!"

	ReportNotFound = "گزارشی یافت نشد"
)
