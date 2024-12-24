package forms

import (
	"github.com/pro-assistance-dev/sprob/modules/forms/handlers/answervariants"
	"github.com/pro-assistance-dev/sprob/modules/forms/handlers/fieldfills"
	"github.com/pro-assistance-dev/sprob/modules/forms/handlers/fields"
	"github.com/pro-assistance-dev/sprob/modules/forms/handlers/formfills"
	"github.com/pro-assistance-dev/sprob/modules/forms/handlers/forms"
	"github.com/pro-assistance-dev/sprob/modules/forms/handlers/formsections"
	"github.com/pro-assistance-dev/sprob/modules/forms/handlers/selectedanswervariants"

	answervariantsRouter "github.com/pro-assistance-dev/sprob/modules/forms/routing/answervariants"
	fieldfillsRouter "github.com/pro-assistance-dev/sprob/modules/forms/routing/fieldfills"
	fieldsRouter "github.com/pro-assistance-dev/sprob/modules/forms/routing/fields"
	formFillsRouter "github.com/pro-assistance-dev/sprob/modules/forms/routing/formfills"
	formsRouter "github.com/pro-assistance-dev/sprob/modules/forms/routing/forms"
	formsectionsRouter "github.com/pro-assistance-dev/sprob/modules/forms/routing/formsections"
	selectedanswervariantsRouter "github.com/pro-assistance-dev/sprob/modules/forms/routing/selectedanswervariants"

	"github.com/gin-gonic/gin"
	helperPack "github.com/pro-assistance-dev/sprob/helper"
)

func InitRoutes(api *gin.RouterGroup, helper *helperPack.Helper) {
	formsections.Init(helper)
	formsectionsRouter.Init(api.Group("/form-sections"), formsections.H)

	forms.Init(helper)
	formsRouter.Init(api.Group("/forms"), forms.H)

	fields.Init(helper)
	fieldsRouter.Init(api.Group("/fields"), fields.H)

	fieldfills.Init(helper)
	fieldfillsRouter.Init(api.Group("/field-fills"), fieldfills.H)

	answervariants.Init(helper)
	answervariantsRouter.Init(api.Group("/answer-variants"), answervariants.H)

	formfills.Init(helper)
	formFillsRouter.Init(api.Group("/form-fills"), formfills.H)

	selectedanswervariants.Init(helper)
	selectedanswervariantsRouter.Init(api.Group("/selected-answer-variants"), selectedanswervariants.H)
}
