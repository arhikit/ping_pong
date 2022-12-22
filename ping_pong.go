package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type score struct {
	points     map[int]int
	nextPlayer int
	mu         sync.Mutex
}

func incPoints(score *score, winner int) {
	score.mu.Lock()
	score.points[winner]++
	score.nextPlayer = winner
	score.mu.Unlock()

	fmt.Printf("Stop. Player %v won. Score %v:%v.\n\n", winner, score.points[1], score.points[2])
	time.Sleep(1 * time.Second)
}

type paramsPingPong struct {
	soundPlayer string
	in          <-chan int
	out         chan<- int
	stop        chan bool
	wg          *sync.WaitGroup
	score       *score
}

func pingPong(params paramsPingPong) {

	for {
		select {

		// прием броска
		case numPlayer := <-params.in:
			fmt.Printf("Player %v: %v\n", numPlayer, params.soundPlayer)

			// с вероятностью 20% игрок может загасить удар
			if rand.Intn(5) == 0 {
				// победитель зарабатывает 1 очко
				incPoints(params.score, numPlayer)

				// переход к следующему раунду
				params.stop <- true
				params.wg.Done()

			} else {
				// игрок отбрасывает мяч сопернику
				nextPlayer := numPlayer%2 + 1
				params.out <- nextPlayer
			}

		//	закрытие потока, тк раунд завершился
		case <-params.stop:
			params.wg.Done()
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	firstChan := make(chan int, 1)
	secondChan := make(chan int, 1)
	stop := make(chan bool)
	wg := &sync.WaitGroup{}
	score := &score{points: map[int]int{1: 0, 2: 0}, nextPlayer: 1}

	for {
		// запуск игры
		fmt.Println("Begin")
		time.Sleep(1 * time.Second)

		// устанавливаем группу ожидания для двух потоков
		wg.Add(2)

		// определяем имя первого игрока
		firstChan <- score.nextPlayer

		// запускаем первого игрока
		paramsPing := paramsPingPong{}
		paramsPing.soundPlayer = "ping"
		paramsPing.in = firstChan
		paramsPing.out = secondChan
		paramsPing.stop = stop
		paramsPing.wg = wg
		paramsPing.score = score
		go pingPong(paramsPing)

		// запускаем второго игрока
		paramsPong := paramsPingPong{}
		paramsPong.soundPlayer = "pong"
		paramsPong.in = secondChan
		paramsPong.out = firstChan
		paramsPong.stop = stop
		paramsPong.wg = wg
		paramsPong.score = score
		go pingPong(paramsPong)

		// ожидание завершения обоих игроков
		wg.Wait()

		// проверка продолжения игры: один из игроков набрал 10 очков
		if score.points[1] == 10 || score.points[2] == 10 {
			break
		}
	}

	// подсчет очков
	titleWinner := "Draw."
	if score.points[1] > score.points[2] {
		titleWinner = "Player 1 won."
	} else if score.points[1] < score.points[2] {
		titleWinner = "Player 2 won."
	}
	fmt.Printf("Score %v:%v. %v\n", score.points[1], score.points[2], titleWinner)

}
