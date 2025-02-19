package messages

const (
	GeneralSuccess      = "عملیات با موفقیت انجام شد"
	GeneralFailed       = "عملیات با شکست مواجه شد"
	GeneralUnauthorized = "برای دسترسی به این بخش لطفا وارد شوید"

	OTPInvalid  = "کد وارد شده صحیح نیست"
	OTPTryAgain = "لطفا بعد از گذشت %d ثانیه مجددا تلاش کنید"

	StoryNotFound   = "داستانی یافت نشد"
	StoryCreated    = "داستان با موفقیت ثبت شد"
	StoryNotCreated = "ثبت داستان موفقیت آمیز نبود"
	StoryEdited     = "داستان با موفقیت ویرایش شد"
	StoryDeleted    = "داستان با موفقیت حذف شد"
	// StoryMinCharLimit = "داستان باید حداقل شامل ۲۵ حرف باشد"
	// StoryMaxCharLimit = "داستان می‌تواند نهایتا شامل ۲۵۶ حرف باشد"

	UserNotFound      = "کاربری با این شناسه یافت نشد"
	UserEdited        = "اطلاعات شما با موفقیت ویرایش شد"
	UserForbiddenName = "این نام برای انتخاب مجاز نیست"
)
