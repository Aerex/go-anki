package screen

import (
	"fmt"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/aerex/go-anki/api/sql/sqlite/services"
	sched "github.com/aerex/go-anki/api/sql/sqlite/services/sched/v2"
	"github.com/aerex/go-anki/pkg/models"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog"
)

type QuestionAnswerView struct {
	*tview.Box
	ColService     services.ColService
	DeckService    services.DeckService
	SchedService   sched.SchedService
	QAs            []*models.CardQA
	Log            *zerolog.Logger
	markdownRender *md.Converter
	currentQA      int
	showAnswerQA   map[int]bool
	primatives     map[string]tview.Primitive
}

type StudyView struct {
	ColService     services.ColService
	DeckService    services.DeckService
	SchedService   sched.SchedService
	QAs            []*models.CardQA
	Log            *zerolog.Logger
	markdownRender *md.Converter
	currentQA      int
	primatives     map[string]tview.Primitive
	showAnswerQA   map[int]bool
	app            *tview.Application
}

func NewQuestionAnswerView(QAs []*models.CardQA, log *zerolog.Logger,
	ds services.DeckService, ss sched.SchedService, cs services.ColService, p map[string]tview.Primitive) *QuestionAnswerView {
	view := &QuestionAnswerView{
		Box:            tview.NewBox(),
		QAs:            QAs,
		ColService:     cs,
		SchedService:   ss,
		DeckService:    ds,
		markdownRender: md.NewConverter("", true, nil),
		primatives:     p,
		Log:            log,
		showAnswerQA:   make(map[int]bool),
	}
	p["QA"] = view
	view.primatives = p
	return view
}

func NewStudyView(QAs []*models.CardQA, log *zerolog.Logger,
	ds services.DeckService, ss sched.SchedService, cs services.ColService, p map[string]tview.Primitive) *StudyView {
	return &StudyView{
		QAs:            QAs,
		ColService:     cs,
		SchedService:   ss,
		DeckService:    ds,
		markdownRender: md.NewConverter("", true, nil),
		Log:            log,
		app:            tview.NewApplication(),
		primatives:     p,
		showAnswerQA:   make(map[int]bool),
	}
}

func (q *QuestionAnswerView) Draw(screen tcell.Screen) {
	q.DrawForSubclass(screen, q)
	x, y, width, height := q.GetInnerRect()
	q.markdownRender.AddRules(md.Rule{
		Filter: []string{"hr"},
		Replacement: func(content string, sel *goquery.Selection, opt *md.Options) *string {
			breakln := "\\hr"
			return &breakln
		},
	})
	var qaText string
	if q.showAnswerQA[q.currentQA] {
		qaText = q.QAs[q.currentQA].AnswerBrowser
	} else {
		qaText = q.QAs[q.currentQA].QuestionBrowser
	}

	q.Log.Debug().Msg(qaText)

	markdownText, err := q.markdownRender.ConvertString(qaText)
	if err != nil {
		q.Log.Fatal().Err(err).Msgf("failed to convert html to markdown for %s", qaText)
	}
	parts := strings.Split(markdownText, "\\hr")
	if len(parts) > 1 {
		tview.Print(screen, parts[0], x, y+height/4, width, tview.AlignCenter, tcell.ColorWhite)
		hr := strings.Repeat(string(tview.Borders.Horizontal), (x+width)/2)
		tview.Print(screen, hr, x, y+height/2, width, tview.AlignCenter, tcell.ColorWhite)
		tview.Print(screen, parts[1], x, y+height/2+1, width, tview.AlignCenter, tcell.ColorWhite)
	} else {
		tview.Print(screen, markdownText, x, y+height/4, width, tview.AlignCenter, tcell.ColorWhite)
	}
}

func (q *QuestionAnswerView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return q.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Rune() {
		case 'u':
			q.currentQA--
			if q.currentQA < 0 {
				q.currentQA = 0
			}
		}
		switch event.Key() {
		case tcell.KeyEnter:
			q.showAnswerQA[q.currentQA] = true
			setFocus(q.primatives["ease"])
		}
	})
}

func (q *StudyView) btmNavBar() *tview.Flex {
	view := tview.NewFlex()
	q.primatives["btmNav"] = view
	view.SetDirection(tview.FlexRowCSS)
	editNoteBtn := tview.NewTextView().SetText("Edit Note")
	editTags := tview.NewTextView().SetText("Edit Tags")
	buryCard := tview.NewTextView().SetText("Bury")
	suspend := tview.NewTextView().SetText("Suspend")
	deleteBtn := tview.NewTextView().SetText("Delete")
	markBtn := tview.NewTextView().SetText("Mark")
	resche := tview.NewTextView().SetText("Reschedule")

	view.AddItem(editNoteBtn, 0, 1, false)
	view.AddItem(editTags, 0, 1, false)
	view.AddItem(buryCard, 0, 1, false)
	view.AddItem(suspend, 0, 1, false)
	view.AddItem(deleteBtn, 0, 1, false)
	view.AddItem(markBtn, 0, 1, false)
	view.AddItem(resche, 0, 1, false)

	return view
}

func (q *StudyView) sessionInfo(deckName string, stats models.DeckStudyStats) *tview.Flex {
	view := tview.NewFlex()
	q.primatives["session"] = view
	view.SetDirection(tview.FlexColumnCSS)
	deckNameView := tview.NewTextView().SetText(deckName).SetTextAlign(tview.AlignLeft)

	// learn - blue; review - red; new - green
	statsView := tview.NewTextView()
	fmt.Fprintf(statsView, "[#0000ff]%d [#ff0000::u]%d[-:-:-] [#00ff00]%d", stats.Learning, stats.Review, stats.New)
	statsView.SetTextAlign(tview.AlignLeft).SetDynamicColors(true)

	view.AddItem(deckNameView, 1, 2, false)
	view.AddItem(statsView, 1, 2, false)

	return view
}

func (q *StudyView) rightNav() *tview.Flex {
	view := tview.NewFlex()
	q.primatives["rightNav"] = view
	view.SetDirection(tview.FlexColumnCSS)

	undo := tview.NewTextView().SetText(" [::u]U[-:-:-]ndo").SetTextAlign(tview.AlignRight).SetDynamicColors(true)
	view.AddItem(undo, 0, 1, false)

	return view
}

func (q *StudyView) easeButtonTimes(ease models.Ease) (string, error) {
	conf, err := q.ColService.Conf()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve collection config")
	}

	if !conf.EstimateTimes {
		return "", nil
	}
	qa := q.QAs[q.currentQA]
	deckConfig, err := q.DeckService.Conf(qa.Card.DeckID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve deck config for %v", qa.Card.DeckID)
	}
	easeTimes, err := q.SchedService.NextIntervalString(qa.Card, ease, deckConfig)
	if err != nil {
		return "", fmt.Errorf("failed to get ease interval time for ease %d on card %d", ease, qa.Card.ID)
	}
	return easeTimes, nil
}

func (q *StudyView) easeButton(cnt int, ease models.Ease, label string) *tview.TextView {
	easeTime, err := q.easeButtonTimes(ease)
	if err != nil {
		q.Log.Fatal().Err(err)
	}
	return tview.NewTextView().SetText(fmt.Sprintf("[%d] %s %s", cnt, label, easeTime)).SetTextAlign(tview.AlignCenter)
}

func (q *StudyView) easeButtonsNav() *tview.Flex {
	view := tview.NewFlex()
	q.primatives["ease"] = view
	view.SetDirection(tview.FlexRowCSS)
	qa := q.QAs[q.currentQA]
	cnt, err := q.SchedService.AnswerButtons(qa.Card)
	if err != nil {
		q.Log.Fatal().Err(err).Msgf("failed to retrieve num of buttons to show")
	}

	againBtn := q.easeButton(1, models.LearnEaseWrong, "Again").SetTextAlign(tview.AlignCenter)
	goodBtn := q.easeButton(2, models.LearnEaseOK, "Good").SetTextAlign(tview.AlignCenter)
	easyBtn := q.easeButton(3, models.LearnEaseEasy, "Easy").SetTextAlign(tview.AlignCenter)
	hardBtn := q.easeButton(4, models.ReviewEaseHard, "Hard").SetTextAlign(tview.AlignCenter)

	view.AddItem(againBtn, 0, 1, false)
	switch cnt {
	case 2:
		view.AddItem(goodBtn, 0, 1, false)
	case 3:
		view.AddItem(goodBtn, 0, 1, false)
		view.AddItem(easyBtn, 0, 1, false)
	default:
		view.AddItem(goodBtn, 0, 1, false)
		view.AddItem(easyBtn, 0, 1, false)
		view.AddItem(hardBtn, 0, 1, false)
	}

	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		var err error
		switch event.Rune() {
		case '1':
			err = q.SchedService.AnswerCard(qa.Card, models.LearnEaseWrong)
		case '2':
			err = q.SchedService.AnswerCard(qa.Card, models.LearnEaseOK)
		case '3':
			err = q.SchedService.AnswerCard(qa.Card, models.LearnEaseEasy)
		case '4':
			err = q.SchedService.AnswerCard(qa.Card, models.ReviewEaseEasy)
		}
		if err != nil {
			q.Log.Fatal().Err(err).Msgf("failed to answer card %v", qa.Card.ID)
		}
		q.currentQA++
		if q.currentQA > len(q.QAs) {
			q.currentQA = len(q.QAs) - 1
		}
		return event
	})
	return view
}

// StudyReview will create a terminal app for studying cards
func StudyReview(log *zerolog.Logger, deckName string, cards []*models.CardQA, stats models.DeckStudyStats,
	schedService sched.SchedService, deckService services.DeckService, colService services.ColService) error {
	primatives := make(map[string]tview.Primitive)
	app := NewStudyView(cards, log, deckService, schedService, colService, primatives)

	qaView := NewQuestionAnswerView(cards, log, deckService, schedService, colService, primatives)

	container := tview.NewGrid().SetColumns(20, 0, 20).SetRows(3, 0, 3, 3)
	container.AddItem(app.sessionInfo(deckName, stats), 0, 0, 1, 1, 0, 0, false)
	container.AddItem(app.rightNav(), 0, 2, 1, 1, 0, 0, false)
	container.AddItem(app.easeButtonsNav(), 2, 1, 1, 1, 0, 0, false)
	container.AddItem(qaView, 1, 0, 1, 3, 0, 0, true)
	container.AddItem(app.btmNavBar(), 3, 1, 1, 1, 0, 0, false)

	if err := app.app.SetRoot(container, true).SetFocus(container).Run(); err != nil {
		return err
	}

	return nil
}
