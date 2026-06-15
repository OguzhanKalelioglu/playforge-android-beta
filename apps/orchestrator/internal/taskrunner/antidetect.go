package taskrunner

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/testerscommunity/orchestrator/internal/task"
)

// AntiDetect, gesture humanization için helper
// Gaussian distribution ile rastgele ama gerçekçi süreler
type AntiDetect struct {
	rng *rand.Rand
}

func NewAntiDetect() *AntiDetect {
	return &AntiDetect{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Delay, aksiyonlar arası bekleme (gaussian μ=1.2s σ=0.4s, clamp [0.3, 4.0]s)
func (a *AntiDetect) Delay(ctx context.Context) time.Duration {
	d := a.gaussian(1200, 400)
	d = clamp(d, 300, 4000)
	timer := time.NewTimer(time.Duration(d) * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return 0
	case <-timer.C:
		return time.Duration(d) * time.Millisecond
	}
}

// GestureDelay, mikro bekleme (μ=80ms σ=30ms, clamp [20, 200]ms)
func (a *AntiDetect) GestureDelay() time.Duration {
	d := a.gaussian(80, 30)
	d = clamp(d, 20, 200)
	time.Sleep(time.Duration(d) * time.Millisecond)
	return time.Duration(d) * time.Millisecond
}

// JitterXY, koordinatı ±6 piksel oynatır
func (a *AntiDetect) JitterXY(x, y int) (int, int) {
	return x + a.rng.Intn(13) - 6, y + a.rng.Intn(13) - 6
}

// SwipePath, x1,y1 → x2,y2 arasında Bezier eğrisi
// n: ara nokta sayısı (default 12)
func (a *AntiDetect) SwipePath(x1, y1, x2, y2, n int) []task.Point {
	if n < 2 {
		n = 12
	}
	points := make([]task.Point, n)

	// 2 kontrol noktası (rastgele perturbation ile yumuşak eğri)
	cx1 := x1 + (x2-x1)/3 + a.rng.Intn(20) - 10
	cy1 := y1 + (y2-y1)/3 + a.rng.Intn(20) - 10
	cx2 := x1 + 2*(x2-x1)/3 + a.rng.Intn(20) - 10
	cy2 := y1 + 2*(y2-y1)/3 + a.rng.Intn(20) - 10

	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1)
		// Cubic Bezier
		invT := 1 - t
		x := invT*invT*invT*float64(x1) +
			3*invT*invT*t*float64(cx1) +
			3*invT*t*t*float64(cx2) +
			t*t*t*float64(x2)
		y := invT*invT*invT*float64(y1) +
			3*invT*invT*t*float64(cy1) +
			3*invT*t*t*float64(cy2) +
			t*t*t*float64(y2)
		// Jitter
		jx, jy := a.JitterXY(int(x), int(y))
		points[i] = task.Point{X: jx, Y: jy}
	}
	return points
}

// AppLaunchPause, app launch sonrası bekleme (μ=2.5s σ=0.6s, clamp [1.8, 3.5]s)
func (a *AntiDetect) AppLaunchPause() time.Duration {
	d := a.gaussian(2500, 600)
	d = clamp(d, 1800, 3500)
	time.Sleep(time.Duration(d) * time.Millisecond)
	return time.Duration(d) * time.Millisecond
}

// EngagementDuration, 2-5dk arası rastgele süre
func (a *AntiDetect) EngagementDuration(minSec, maxSec int) time.Duration {
	if minSec >= maxSec {
		return time.Duration(minSec) * time.Second
	}
	// Gaussian distribution etrafında rastgele
	mid := float64(minSec+maxSec) / 2
	sigma := float64(maxSec-minSec) / 4
	d := a.gaussian(int(mid), int(sigma))
	d = clamp(d, float64(minSec), float64(maxSec))
	return time.Duration(d) * 1000 * time.Millisecond
}

// gaussian, Box-Muller transform ile normal dağılım
func (a *AntiDetect) gaussian(mean, stdDev int) float64 {
	u1 := a.rng.Float64()
	u2 := a.rng.Float64()
	z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
	return float64(mean) + z*float64(stdDev)
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
