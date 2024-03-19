package main

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/montanaflynn/stats"
)

//nolint:gocognit
func main() {
	routerFilterRegex := regexp.MustCompile(`.*Router filter processing completed.*timeTakenMS=(\d+)`)
	virtualServiceRegex := regexp.MustCompile(`.*Virtual service processing completed.*timeTakenMS=(\d+)`)
	throttleFilterRegex := regexp.MustCompile(`.*Throttle filter processing completed.*timeTakenMS=(\d+)`)
	tcTotalProcessingRegex := regexp.MustCompile(`.*Processing completed.*controller=traffic-config-controller.*timeTakenMS=(\d+)`)
	deployTotalProcessingRegex := regexp.MustCompile(`.*Processing completed.*controller=deployment-controller.*timeTakenMS=(\d+)`)
	rolloutTotalProcessingRegex := regexp.MustCompile(`.*Processing completed.*controller=rollouts-controller.*timeTakenMS=(\d+)`)
	serviceTotalProcessingRegex := regexp.MustCompile(`.*Processing completed.*controller=service-controller.*timeTakenMS=(\d+)`)
	dependencyTotalProcessingRegex := regexp.MustCompile(`.*Processing completed.*controller=dependency-controller.*timeTakenMS=(\d+)`)
	trafficConfigQtimeLenRegex := regexp.MustCompile(`.*Processing started. controller=traffic-config-controller.*queueLen=(\d+) queueTimeMS=(\d+)`)
	deploymentQtimeLenRegex := regexp.MustCompile(`.*Processing started. controller=deployment-controller.*queueLen=(\d+) queueTimeMS=(\d+)`)
	rolloutQtimeLenRegex := regexp.MustCompile(`.*Processing started. controller=rollouts-controller.*queueLen=(\d+) queueTimeMS=(\d+)`)
	serviceQtimeLenRegex := regexp.MustCompile(`.*Processing started. controller=service-controller.*queueLen=(\d+) queueTimeMS=(\d+)`)
	dependencyQtimeLenRegex := regexp.MustCompile(`.*Processing started. controller=dependency-controller.*queueLen=(\d+) queueTimeMS=(\d+)`)
	perfSetupRegex := regexp.MustCompile(`.*caches are getting synced.*clusters=(\d+)/.* deployments=(\d+)/.* rollouts=(\d+)/.* depedencyrecord=(\d+)/.*`)

	throttleFilterTimeTaken := []int{}
	virtualServiceTimeTaken := []int{}
	routerFilterTimeTaken := []int{}
	totalProcessingTimeTaken := []int{}
	deployTotalProcessingTimeTaken := []int{}
	rolloutTotalProcessingTimeTaken := []int{}
	serviceTotalProcessingTimeTaken := []int{}
	dependencyTotalProcessingTimeTaken := []int{}
	tcQueueTime := []int{}
	tcQueueLen := []int{}
	deploymentQueueTime := []int{}
	deploymentQueueLen := []int{}
	rolloutQueueTime := []int{}
	rolloutQueueLen := []int{}
	serviceQueueTime := []int{}
	serviceQueueLen := []int{}
	dependencyQueueTime := []int{}
	dependencyQueueLen := []int{}
	totalClusters := 0
	totalDeployments := 0
	totalRollouts := 0
	totalDependencyRecords := 0

	file, err := os.Open("./perf_output.log")
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		routerMatches := routerFilterRegex.FindStringSubmatch(line)
		virtualServiceMatches := virtualServiceRegex.FindStringSubmatch(line)
		throttleFilterMatches := throttleFilterRegex.FindStringSubmatch(line)
		totalProcessingMatches := tcTotalProcessingRegex.FindStringSubmatch(line)
		deployTotalProcessingMatches := deployTotalProcessingRegex.FindStringSubmatch(line)
		rolloutTotalProcessingMatches := rolloutTotalProcessingRegex.FindStringSubmatch(line)
		serviceTotalProcessingMatches := serviceTotalProcessingRegex.FindStringSubmatch(line)
		dependencyTotalProcessingMatches := dependencyTotalProcessingRegex.FindStringSubmatch(line)
		tcQueueTimeAndLenMatches := trafficConfigQtimeLenRegex.FindStringSubmatch(line)
		deploymentQueueTimeAndLenMatches := deploymentQtimeLenRegex.FindStringSubmatch(line)
		rolloutQueueTimeAndLenMatches := rolloutQtimeLenRegex.FindStringSubmatch(line)
		serviceQueueTimeAndLenMatches := serviceQtimeLenRegex.FindStringSubmatch(line)
		dependencyQueueTimeAndLenMatches := dependencyQtimeLenRegex.FindStringSubmatch(line)
		perfSetupMatches := perfSetupRegex.FindStringSubmatch(line)
		if len(routerMatches) > 1 {
			timetaken, _ := strconv.Atoi(routerMatches[1])
			routerFilterTimeTaken = append(routerFilterTimeTaken, timetaken)
			continue
		}
		if len(virtualServiceMatches) > 1 {
			timetaken, _ := strconv.Atoi(virtualServiceMatches[1])
			virtualServiceTimeTaken = append(virtualServiceTimeTaken, timetaken)
			continue
		}
		if len(throttleFilterMatches) > 1 {
			timetaken, _ := strconv.Atoi(throttleFilterMatches[1])
			throttleFilterTimeTaken = append(throttleFilterTimeTaken, timetaken)
			continue
		}
		if len(totalProcessingMatches) > 1 {
			timetaken, _ := strconv.Atoi(totalProcessingMatches[1])
			totalProcessingTimeTaken = append(totalProcessingTimeTaken, timetaken)
			continue
		}
		if len(deployTotalProcessingMatches) > 1 {
			timetaken, _ := strconv.Atoi(deployTotalProcessingMatches[1])
			deployTotalProcessingTimeTaken = append(deployTotalProcessingTimeTaken, timetaken)
			continue
		}
		if len(rolloutTotalProcessingMatches) > 1 {
			timetaken, _ := strconv.Atoi(rolloutTotalProcessingMatches[1])
			rolloutTotalProcessingTimeTaken = append(rolloutTotalProcessingTimeTaken, timetaken)
			continue
		}
		if len(serviceTotalProcessingMatches) > 1 {
			timetaken, _ := strconv.Atoi(serviceTotalProcessingMatches[1])
			serviceTotalProcessingTimeTaken = append(serviceTotalProcessingTimeTaken, timetaken)
			continue
		}
		if len(dependencyTotalProcessingMatches) > 1 {
			timetaken, _ := strconv.Atoi(dependencyTotalProcessingMatches[1])
			dependencyTotalProcessingTimeTaken = append(dependencyTotalProcessingTimeTaken, timetaken)
			continue
		}
		if len(tcQueueTimeAndLenMatches) > 1 {
			queueLen, _ := strconv.Atoi(tcQueueTimeAndLenMatches[1])
			queueTime, _ := strconv.Atoi(tcQueueTimeAndLenMatches[2])
			tcQueueTime = append(tcQueueTime, queueTime)
			tcQueueLen = append(tcQueueLen, queueLen)
			continue
		}
		if len(deploymentQueueTimeAndLenMatches) > 1 {
			queueLen, _ := strconv.Atoi(deploymentQueueTimeAndLenMatches[1])
			queueTime, _ := strconv.Atoi(deploymentQueueTimeAndLenMatches[2])
			deploymentQueueTime = append(deploymentQueueTime, queueTime)
			deploymentQueueLen = append(deploymentQueueLen, queueLen)
			continue
		}
		if len(rolloutQueueTimeAndLenMatches) > 1 {
			queueLen, _ := strconv.Atoi(rolloutQueueTimeAndLenMatches[1])
			queueTime, _ := strconv.Atoi(rolloutQueueTimeAndLenMatches[2])
			rolloutQueueTime = append(rolloutQueueTime, queueTime)
			rolloutQueueLen = append(rolloutQueueLen, queueLen)
			continue
		}
		if len(serviceQueueTimeAndLenMatches) > 1 {
			queueLen, _ := strconv.Atoi(serviceQueueTimeAndLenMatches[1])
			queueTime, _ := strconv.Atoi(serviceQueueTimeAndLenMatches[2])
			serviceQueueTime = append(serviceQueueTime, queueTime)
			serviceQueueLen = append(serviceQueueLen, queueLen)
			continue
		}
		if len(dependencyQueueTimeAndLenMatches) > 1 {
			queueLen, _ := strconv.Atoi(dependencyQueueTimeAndLenMatches[1])
			queueTime, _ := strconv.Atoi(dependencyQueueTimeAndLenMatches[2])
			dependencyQueueTime = append(dependencyQueueTime, queueTime)
			dependencyQueueLen = append(dependencyQueueLen, queueLen)
			continue
		}
		if len(perfSetupMatches) > 1 {
			totalClusters, _ = strconv.Atoi(perfSetupMatches[1])
			totalDeployments, _ = strconv.Atoi(perfSetupMatches[2])
			totalRollouts, _ = strconv.Atoi(perfSetupMatches[3])
			totalDependencyRecords, _ = strconv.Atoi(perfSetupMatches[4])
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	printPerfResults(
		routerFilterTimeTaken,
		virtualServiceTimeTaken,
		throttleFilterTimeTaken,
		totalProcessingTimeTaken,
		deployTotalProcessingTimeTaken,
		rolloutTotalProcessingTimeTaken,
		serviceTotalProcessingTimeTaken,
		dependencyTotalProcessingTimeTaken,
		tcQueueTime,
		tcQueueLen,
		deploymentQueueTime,
		deploymentQueueLen,
		rolloutQueueTime,
		rolloutQueueLen,
		serviceQueueTime,
		serviceQueueLen,
		dependencyQueueTime,
		dependencyQueueLen,
		totalClusters,
		totalDeployments,
		totalRollouts,
		totalDependencyRecords,
	)
	defer file.Close()
}

func printPerfResults(
	routerfilterTimeTaken []int,
	virtualServiceTimeTaken []int,
	throttleFilterTimeTaken []int,
	totalProcessingTimeTaken []int,
	deployTotalProcessingTimeTaken []int,
	rolloutTotalProcessingTimeTaken []int,
	serviceTotalProcessingTimeTaken []int,
	dependencyTotalProcessingTimeTaken []int,
	tcQueueTime []int,
	tcQueueLen []int,
	deploymentQueueTime []int,
	deploymentQueueLen []int,
	rolloutQueueTime []int,
	rolloutQueueLen []int,
	serviceQueueTime []int,
	serviceQueueLen []int,
	dependencyQueueTime []int,
	dependencyQueueLen []int,
	totalClusters int,
	totalDeployments int,
	totalRollouts int,
	totalDependencyRecords int,
) {
	routerfilterTTAvg, routerfilterTT50, routerfilterTT75, routerfilterTT90, routerfilterTT95, routerfilterTT99 := getStats(routerfilterTimeTaken)
	virtualServiceTTAvg, virtualServiceTT50, virtualServiceTT75, virtualServiceTT90, virtualServiceTT95, virtualServiceTT99 := getStats(virtualServiceTimeTaken)
	throttleFilterTTAvg, throttleFilterTT50, throttleFilterTT75, throttleFilterTT90, throttleFilterTT95, throttleFilterTT99 := getStats(throttleFilterTimeTaken)
	totalProcessingTTAvg, totalProcessingTT50, totalProcessingTT75, totalProcessingTT90, totalProcessingTT95, totalProcessingTT99 := getStats(totalProcessingTimeTaken)
	deployTotalProcessingTTAvg, deployTotalProcessingTT50, deployTotalProcessingTT75, deployTotalProcessingTT90, deployTotalProcessingTT95, deployTotalProcessingTT99 := getStats(deployTotalProcessingTimeTaken)
	rolloutTotalProcessingTTAvg, rolloutTotalProcessingTT50, rolloutTotalProcessingTT75, rolloutTotalProcessingTT90, rolloutTotalProcessingTT95, rolloutTotalProcessingTT99 := getStats(rolloutTotalProcessingTimeTaken)
	serviceTotalProcessingTTAvg, serviceTotalProcessingTT50, serviceTotalProcessingTT75, serviceTotalProcessingTT90, serviceTotalProcessingTT95, serviceTotalProcessingTT99 := getStats(serviceTotalProcessingTimeTaken)
	dependencyTotalProcessingTTAvg, dependencyTotalProcessingTT50, dependencyTotalProcessingTT75, dependencyTotalProcessingTT90, dependencyTotalProcessingTT95, dependencyTotalProcessingTT99 := getStats(dependencyTotalProcessingTimeTaken)
	tcQueueTimeAvg, tcQueueTime50, tcQueueTime75, tcQueueTime90, tcQueueTime95, tcQueueTime99 := getStats(tcQueueTime)
	tcQueueLenAvg, tcQueueLen50, tcQueueLen75, tcQueueLen90, tcQueueLen95, tcQueueLen99 := getStats(tcQueueLen)
	deploymentQueueTimeAvg, deploymentQueueTime50, deploymentQueueTime75, deploymentQueueTime90, deploymentQueueTime95, deploymentQueueTime99 := getStats(deploymentQueueTime)
	deploymentQueueLenAvg, deploymentQueueLen50, deploymentQueueLen75, deploymentQueueLen90, deploymentQueueLen95, deploymentQueueLen99 := getStats(deploymentQueueLen)
	rolloutQueueTimeAvg, rolloutQueueTime50, rolloutQueueTime75, rolloutQueueTime90, rolloutQueueTime95, rolloutQueueTime99 := getStats(rolloutQueueTime)
	rolloutQueueLenAvg, rolloutQueueLen50, rolloutQueueLen75, rolloutQueueLen90, rolloutQueueLen95, rolloutQueueLen99 := getStats(rolloutQueueLen)
	serviceQueueTimeAvg, serviceQueueTime50, serviceQueueTime75, serviceQueueTime90, serviceQueueTime95, serviceQueueTime99 := getStats(serviceQueueTime)
	serviceQueueLenAvg, serviceQueueLen50, serviceQueueLen75, serviceQueueLen90, serviceQueueLen95, serviceQueueLen99 := getStats(serviceQueueLen)
	dependencyQueueTimeAvg, dependencyQueueTime50, dependencyQueueTime75, dependencyQueueTime90, dependencyQueueTime95, dependencyQueueTime99 := getStats(dependencyQueueTime)
	dependencyQueueLenAvg, dependencyQueueLen50, dependencyQueueLen75, dependencyQueueLen90, dependencyQueueLen95, dependencyQueueLen99 := getStats(dependencyQueueLen)

	t := table.NewWriter()
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	t.SetOutputMirror(os.Stdout)
	t.AppendRow(table.Row{"Perf Setup", "Count", "Count", "Count", "Count", "Count", "Count", "Count"}, rowConfigAutoMerge)
	t.AppendSeparator()
	t.AppendRow(table.Row{"Total Clusters", totalClusters, totalClusters, totalClusters, totalClusters, totalClusters, totalClusters, totalClusters}, rowConfigAutoMerge)
	t.AppendRow(table.Row{"Total Deployments", totalDeployments, totalDeployments, totalDeployments, totalDeployments, totalDeployments, totalDeployments, totalDeployments}, rowConfigAutoMerge)
	t.AppendRow(table.Row{"Total Rollouts", totalRollouts, totalRollouts, totalRollouts, totalRollouts, totalRollouts, totalRollouts, totalRollouts}, rowConfigAutoMerge)
	t.AppendRow(table.Row{"Total Dependency Records", totalDependencyRecords, totalDependencyRecords, totalDependencyRecords, totalDependencyRecords, totalDependencyRecords, totalDependencyRecords, totalDependencyRecords}, rowConfigAutoMerge)

	t.AppendSeparator()
	t.AppendRow(table.Row{"Naavik Performance Results", "Time Taken (milliseconds)", "Time Taken (milliseconds)", "Time Taken (milliseconds)", "Time Taken (milliseconds)", "Time Taken (milliseconds)", "Time Taken (milliseconds)", "Time Taken (milliseconds)"}, rowConfigAutoMerge)
	t.AppendSeparator()
	t.AppendRow(table.Row{"Name", "Total Results", "Avg", "p50", "p75", "p90", "p95", "p99"})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Deployment informer queue time", len(deploymentQueueTime), deploymentQueueTimeAvg, deploymentQueueTime50, deploymentQueueTime75, deploymentQueueTime90, deploymentQueueTime95, deploymentQueueTime99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Deployment informer queue len", len(deploymentQueueLen), deploymentQueueLenAvg, deploymentQueueLen50, deploymentQueueLen75, deploymentQueueLen90, deploymentQueueLen95, deploymentQueueLen99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Deployment controller processing", len(deployTotalProcessingTimeTaken), deployTotalProcessingTTAvg, deployTotalProcessingTT50, deployTotalProcessingTT75, deployTotalProcessingTT90, deployTotalProcessingTT95, deployTotalProcessingTT99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"", "", "", "", "", "", "", ""}, rowConfigAutoMerge)
	t.AppendSeparator()
	t.AppendRow(table.Row{"Rollout informer queue time", len(rolloutQueueTime), rolloutQueueTimeAvg, rolloutQueueTime50, rolloutQueueTime75, rolloutQueueTime90, rolloutQueueTime95, rolloutQueueTime99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Rollout informer queue len", len(rolloutQueueLen), rolloutQueueLenAvg, rolloutQueueLen50, rolloutQueueLen75, rolloutQueueLen90, rolloutQueueLen95, rolloutQueueLen99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Rollout controller processing", len(rolloutTotalProcessingTimeTaken), rolloutTotalProcessingTTAvg, rolloutTotalProcessingTT50, rolloutTotalProcessingTT75, rolloutTotalProcessingTT90, rolloutTotalProcessingTT95, rolloutTotalProcessingTT99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"", "", "", "", "", "", "", ""}, rowConfigAutoMerge)
	t.AppendSeparator()
	t.AppendRow(table.Row{"Service informer queue time", len(serviceQueueTime), serviceQueueTimeAvg, serviceQueueTime50, serviceQueueTime75, serviceQueueTime90, serviceQueueTime95, serviceQueueTime99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Service informer queue len", len(serviceQueueLen), serviceQueueLenAvg, serviceQueueLen50, serviceQueueLen75, serviceQueueLen90, serviceQueueLen95, serviceQueueLen99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Service controller processing", len(serviceTotalProcessingTimeTaken), serviceTotalProcessingTTAvg, serviceTotalProcessingTT50, serviceTotalProcessingTT75, serviceTotalProcessingTT90, serviceTotalProcessingTT95, serviceTotalProcessingTT99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"", "", "", "", "", "", "", ""}, rowConfigAutoMerge)
	t.AppendSeparator()
	t.AppendRow(table.Row{"Dependency informer queue time", len(dependencyQueueTime), dependencyQueueTimeAvg, dependencyQueueTime50, dependencyQueueTime75, dependencyQueueTime90, dependencyQueueTime95, dependencyQueueTime99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Dependency informer queue len", len(dependencyQueueLen), dependencyQueueLenAvg, dependencyQueueLen50, dependencyQueueLen75, dependencyQueueLen90, dependencyQueueLen95, dependencyQueueLen99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Dependency controller processing", len(dependencyTotalProcessingTimeTaken), dependencyTotalProcessingTTAvg, dependencyTotalProcessingTT50, dependencyTotalProcessingTT75, dependencyTotalProcessingTT90, dependencyTotalProcessingTT95, dependencyTotalProcessingTT99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"", "", "", "", "", "", "", ""}, rowConfigAutoMerge)
	t.AppendSeparator()
	t.AppendRow(table.Row{"TrafficConfig informer queue time", len(tcQueueTime), tcQueueTimeAvg, tcQueueLen50, tcQueueLen75, tcQueueTime90, tcQueueTime95, tcQueueTime99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"TrafficConfig informer queue len", len(tcQueueLen), tcQueueLenAvg, tcQueueTime50, tcQueueTime75, tcQueueLen90, tcQueueLen95, tcQueueLen99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Router filter gen", len(routerfilterTimeTaken), routerfilterTTAvg, routerfilterTT50, routerfilterTT75, routerfilterTT90, routerfilterTT95, routerfilterTT99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Virtual service gen", len(virtualServiceTimeTaken), virtualServiceTTAvg, virtualServiceTT50, virtualServiceTT75, virtualServiceTT90, virtualServiceTT95, virtualServiceTT99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Throttle filter gen", len(throttleFilterTimeTaken), throttleFilterTTAvg, throttleFilterTT50, throttleFilterTT75, throttleFilterTT90, throttleFilterTT95, throttleFilterTT99})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Total processing", len(totalProcessingTimeTaken), totalProcessingTTAvg, totalProcessingTT50, totalProcessingTT75, totalProcessingTT90, totalProcessingTT95, totalProcessingTT99})

	t.Render()
}

func getStats(data []int) (avg, p50, p75, p90, p95, p99 float64) {
	statsData := stats.LoadRawData(data)
	avg, _ = statsData.Mean()
	avg, _ = stats.Round(avg, 2)
	p50, _ = statsData.Percentile(50)
	p75, _ = statsData.Percentile(75)
	p90, _ = statsData.Percentile(90)
	p95, _ = statsData.Percentile(95)
	p99, _ = statsData.Percentile(99)
	return
}
