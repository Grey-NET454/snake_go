package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Позиция на игровом поле
type Position struct {
	X, Y int
}

// Состояние игры
type Game struct {
	snake     []Position
	direction Position
	nextDir   Position
	food      Position
	score     int
	gameOver  bool
	width     int
	height    int
}

// Инициализация игры
func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	g := &Game{
		direction: Position{1, 0}, // Начальное движение вправо
		nextDir:   Position{1, 0},
		score:     0,
		width:     40,
		height:    20,
		gameOver:  false,
	}
	g.reset()
	return g
}

// Сброс состояния игры
func (g *Game) reset() {
	// Начальная змейка из 3 сегментов
	g.snake = []Position{
		{X: 5, Y: 10},
		{X: 4, Y: 10},
		{X: 3, Y: 10},
	}
	g.direction = Position{1, 0}
	g.nextDir = Position{1, 0}
	g.score = 0
	g.gameOver = false
	g.placeFood()
}

// Размещение еды в случайном месте
func (g *Game) placeFood() {
	for {
		x := rand.Intn(g.width)
		y := rand.Intn(g.height)
		// Проверка, что еда не на змейке
		onSnake := false
		for _, seg := range g.snake {
			if seg.X == x && seg.Y == y {
				onSnake = true
				break
			}
		}
		if !onSnake {
			g.food = Position{X: x, Y: y}
			return
		}
	}
}

// Обновление состояния игры
func (g *Game) update() {
	if g.gameOver {
		return
	}

	g.direction = g.nextDir

	// Новая голова
	head := g.snake[0]
	head.X += g.direction.X
	head.Y += g.direction.Y

	// Проверка столкновения со стенами
	if head.X < 0 || head.X >= g.width || head.Y < 1 || head.Y > g.height {
		g.gameOver = true
		return
	}

	// Проверка столкновения с собой
	for _, seg := range g.snake {
		if seg.X == head.X && seg.Y == head.Y {
			g.gameOver = true
			return
		}
	}

	// Добавляем новую голову
	g.snake = append([]Position{head}, g.snake...)

	// Проверка, съела ли еду
	if head.X == g.food.X && head.Y == g.food.Y {
		g.score++
		g.placeFood()
	} else {
		// Удаляем хвост, если не съела еду
		g.snake = g.snake[:len(g.snake)-1]
	}
}

// Обработка ввода
func (g *Game) handleInput(key string) {
	switch key {
	case "w", "W":
		if g.direction.Y != 1 {
			g.nextDir = Position{0, -1}
		}
	case "s", "S":
		if g.direction.Y != -1 {
			g.nextDir = Position{0, 1}
		}
	case "a", "A":
		if g.direction.X != 1 {
			g.nextDir = Position{-1, 0}
		}
	case "d", "D":
		if g.direction.X != -1 {
			g.nextDir = Position{1, 0}
		}
	}
}

// Отрисовка игрового поля в буфер строк
func (g *Game) render(screen []string) {
	// Заголовок
	screen[0] = fmt.Sprintf("Змейка | Счёт: %d | Q - выход", g.score)

	// Границы
	horizontal := "+"
	for i := 0; i < g.width; i++ {
		horizontal += "-"
	}
	horizontal += "+"

	// Игровое поле
	for y := 1; y <= g.height; y++ {
		line := "|"
		for x := 0; x < g.width; x++ {
			char := " "
			// Рисуем змейку
			for i, seg := range g.snake {
				if seg.X == x && seg.Y == y {
					if i == 0 {
						char = "@" // Голова
					} else {
						char = "o" // Тело
					}
					break
				}
			}
			// Рисуем еду
			if g.food.X == x && g.food.Y == y {
				char = "*"
			}
			line += char
		}
		line += "|"
		screen[y] = line
	}

	// Нижняя граница
	screen[g.height] = horizontal

	// Нижние строки
	if g.gameOver {
		screen[g.height+1] = "ИГРА ОКОНЧЕНА!"
		screen[g.height+2] = "Нажмите R для рестарта, Q для выхода"
	} else {
		screen[g.height+1] = "Управление: W/S/A/D, выход - Q"
		screen[g.height+2] = ""
	}
}

func main() {
	game := NewGame()
	screenSize := game.height + 3
	screen := make([]string, screenSize)

	fmt.Println("=================================")
	fmt.Println("     Змейка (консольная)        ")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Управление:")
	fmt.Println("  W - вверх")
	fmt.Println("  S - вниз")
	fmt.Println("  A - влево")
	fmt.Println("  D - вправо")
	fmt.Println("  Q - выход")
	fmt.Println()
	fmt.Println("Нажмите Enter для начала игры...")

	var start string
	fmt.Scanln(&start)

	// Скрываем курсор и включаем raw-режим чтения клавиатуры
	if err := enableRawMode(); err != nil {
		fmt.Println("Не удалось включить режим чтения клавиатуры:", err)
		fmt.Println("Запустите игру в терминале, поддерживающем чтение клавиш (не перенаправленный ввод).")
		return
	}
	defer disableRawMode()

	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	inputCh := make(chan byte, 16)
	go readKeys(inputCh)

	printHelp := true

	for {
		// Неблокирующая обработка ввода
		for {
			select {
			case key := <-inputCh:
				switch key {
				case 'q', 'Q':
					fmt.Println("\nВыход из игры. Итоговый счёт:", game.score)
					return
				case 'w', 'W', 's', 'S', 'a', 'A', 'd', 'D':
					game.handleInput(string(key))
				case 'r', 'R':
					if game.gameOver {
						game.reset()
						printHelp = true
					}
				}
			default:
				goto next
			}
		}
	next:
		select {
		case <-ticker.C:
			game.update()

			game.render(screen)

			// Вывод с возвратом каретки в начало (двойная буферизация)
			if printHelp {
				// Прокрутка вверх на размер экрана
				fmt.Print("\033[2J\033[H")
				printHelp = false
			}
			fmt.Print("\033[H")
			for _, line := range screen {
				fmt.Println(line)
			}

			if game.gameOver {
				// После окончания игры замедляемся, ждём R или Q
				ticker.Reset(500 * time.Millisecond)
			} else {
				ticker.Reset(150 * time.Millisecond)
			}
		}
	}
}

// Очистка терминала
func clearTerminal() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}
