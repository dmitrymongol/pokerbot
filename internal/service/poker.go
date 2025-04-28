package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type HandHistory struct {
	HandID              string
	TournamentID        string
	SmallBlind         int
	BigBlind           int
	Ante               int
	BlindIncreaseInterval time.Duration
	Players            []Player
	Actions            []Action
	Community          []string
	Winners            []string
	MysteryElements    []string 
}

type Player struct {
	Name  string
	Stack int
}

type Action struct {
	Street string // "HOLE CARDS", "FLOP", "TURN", "RIVER"
	Player string // Имя игрока
	Move   string // "raises", "calls", "folds", "checks"
	Amount int    // Размер ставки
}

var (
    ErrInvalidStructure = errors.New("invalid tournament structure")
    ErrBettingError     = errors.New("betting rules violation")
)

func ParseTextHandHistory(text string) (*HandHistory, error) {
    hh := &HandHistory{}
    
    // Новое регулярное выражение для ID
    re := regexp.MustCompile(`(?mi)^Hand #(\d+).*Tournament #([^\s]+)`)
    matches := re.FindStringSubmatch(text)
    if len(matches) >= 3 {
        hh.HandID = matches[1]
        hh.TournamentID = matches[2]
        hh.TournamentID = strings.TrimSuffix(hh.TournamentID, "-") // Удаляем лишние символы
    }

    // Добавляем парсинг Mystery Elements
    hh.parseMysteryElements(text) // <-- Вызов метода

    // Остальной парсинг...
    hh.ParseBlinds(text)
    hh.ParseBlindIncrease(text)

    // Парсинг действий
    actionRe := regexp.MustCompile(`(?mi)^(\w+)\s+(\w+)\s+(to\s+)?([\d,]+)?$`)
    for _, line := range strings.Split(text, "\n") {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "*** ") && strings.HasSuffix(line, " ***") {
            hh.Actions = append(hh.Actions, Action{
                Street: strings.Trim(line, "* "),
            })
            continue
        }

        if matches := actionRe.FindStringSubmatch(line); matches != nil {
            action := Action{
                Player: matches[1],
                Move:   matches[2],
            }

            if matches[4] != "" {
                action.Amount = parseInt(matches[4])
            }

            hh.Actions = append(hh.Actions, action)
        }
    }

    return hh, nil
}

func parseInt(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// Добавляем парсинг блайндов и интервала
func (hh *HandHistory) ParseBlinds(text string) error {
	// Парсим структуру блайндов: "Blinds: 1,000/2,000 (Ante 200)"
	re := regexp.MustCompile(`Blinds: ([\d,]+)/([\d,]+).*Ante ([\d,]+)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) == 4 {
		hh.SmallBlind = parseInt(matches[1])
		hh.BigBlind = parseInt(matches[2])
		hh.Ante = parseInt(matches[3])
		return nil
	}
	return errors.New("failed to parse blinds structure")
}

// Парсим интервал повышения блайндов
func (hh *HandHistory) ParseBlindIncrease(text string) error {
	re := regexp.MustCompile(`Blind Levels: ([\d]+) minutes`)
	if matches := re.FindStringSubmatch(text); len(matches) > 0 {
		minutes, _ := time.ParseDuration(matches[1] + "m")
		hh.BlindIncreaseInterval = minutes
		return nil
	}
	return errors.New("failed to parse blind increase interval")
}

// Проверка Mystery-элементов
func hasMysteryElements(hh *HandHistory) bool {
    // Проверяем наличие любых элементов, специфичных для формата
    requiredElements := []string{"Bounty", "Progressive", "Mystery", "Boost"}
    for _, el := range hh.MysteryElements {
        for _, keyword := range requiredElements {
            if strings.Contains(strings.ToLower(el), strings.ToLower(keyword)) {
                return true
            }
        }
    }
    return len(hh.MysteryElements) > 0
}

func ValidateMysteryRoyale(hh *HandHistory) []error {
	var errs []error

	// 1. Проверка структуры анте
	if hh.Ante < 100 || hh.Ante > hh.BigBlind/2 {
		errs = append(errs, errors.New("invalid ante size"))
	}

	// // 2. Проверка скорости роста блайндов
	// if hh.BlindIncreaseInterval < 5*time.Minute {
	// 	errs = append(errs, errors.New("blind increase interval too short"))
	// }

	// 3. Проверка Mystery-условий
	if !hasMysteryElements(hh) {
		errs = append(errs, errors.New("missing mystery elements"))
	}

	// 4. Дополнительные проверки для Mystery Battle Royale
	if hh.BigBlind < 2000 || hh.BigBlind > 100000 {
		errs = append(errs, errors.New("invalid big blind size for mystery format"))
	}

	return errs
}

func (hh *HandHistory) parseMysteryElements(text string) {
    re := regexp.MustCompile(`(?i)Mystery Elements:\s*\[([^\]]+)\]`)
    if matches := re.FindStringSubmatch(text); len(matches) > 1 {
        elements := strings.Split(matches[1], ", ")
        for i, el := range elements {
            elements[i] = strings.TrimSpace(el)
        }
        hh.MysteryElements = elements
    }
}