package integration

import (
	"context"
	"fmt"
	"sort"
	"testing"

	i18n "forge.lthn.ai/core/go-i18n"
	"forge.lthn.ai/core/go-inference"
	_ "forge.lthn.ai/core/go-mlx" // registers Metal backend
)

// buildCalibrationCorpus constructs 500 samples for 1B vs 27B comparison.
// First 220 have ground truth (from the classification benchmark), the rest
// are diverse prompts without labels for agreement-only measurement.
func buildCalibrationCorpus() []i18n.CalibrationSample {
	var samples []i18n.CalibrationSample

	// --- Ground truth samples (220): 55 per domain ---

	technical := []string{
		"Delete the configuration file",
		"Build the project from source",
		"Run the tests before committing",
		"Push the changes to the branch",
		"Update the dependencies",
		"Check the build status",
		"Find the failing test",
		"Write the test cases first",
		"Set the environment variables",
		"Split the package into modules",
		"Scan the repository for vulnerabilities",
		"Format the source files",
		"Reset the branch to the previous commit",
		"Stop the running process",
		"Cut a new release branch",
		"Send the build artifacts to the server",
		"Keep the test coverage above the threshold",
		"Hold the deployment until the checks pass",
		"Begin the migration to the new package",
		"Take the old server offline",
		"The build failed because of a missing dependency",
		"The test committed changes to the wrong branch",
		"We found a vulnerability in the package",
		"The commit broke the build",
		"She deleted the old configuration files",
		"They pushed the fix to the repository",
		"The branch was updated with the latest changes",
		"He rebuilt the project after updating dependencies",
		"The task failed during the scanning phase",
		"We split the repository into separate packages",
		"The check ran successfully on all branches",
		"They found the issue in the build directory",
		"The file was committed without running tests",
		"Merge the pull request after review",
		"Deploy the service to the staging cluster",
		"Revert the last three commits",
		"Enable verbose logging for debugging",
		"Pin the dependency to version two",
		"Rotate the API keys on production",
		"Profile the memory usage under load",
		"Containerise the application with Docker",
		"Migrate the database schema to version five",
		"Monitor the error rate after deployment",
		"Invalidate the CDN cache for the assets",
		"The pipeline timed out on the integration step",
		"Rollback failed because the snapshot was corrupted",
		"The linter caught twelve style violations",
		"Cache invalidation caused stale data in staging",
		"The DNS propagation took longer than expected",
		"Thread pool exhaustion under concurrent requests",
		"The certificate expired and TLS handshakes failed",
		"Garbage collection pauses exceeded the SLA threshold",
		"Hot-reload broke after upgrading the framework",
		"The socket connection was reset by the load balancer",
		"Rate limiting kicked in after the traffic spike",
	}

	creative := []string{
		"She wrote the story by candlelight",
		"He drew a map of forgotten places",
		"The river froze under the winter moon",
		"They sang the old songs by the fire",
		"She found a letter hidden in the pages",
		"He carved the figure from driftwood",
		"The wind spoke through the hollow trees",
		"They wove the colours into the tapestry",
		"She built a castle from the broken stones",
		"He told the tale of the sunken ship",
		"She painted the sky with broad red strokes",
		"He composed the melody in a single night",
		"They danced beneath the flickering lanterns",
		"The cat sat on the manuscript and purred",
		"She folded the paper into a paper crane",
		"He read the poem aloud to the empty room",
		"They carved their names into the old oak tree",
		"She spun the yarn into a glowing thread",
		"He wrote the first line and then stopped",
		"The garden grew wild after the artist left",
		"Write a ballad about the last lighthouse keeper",
		"Describe the colour of silence at midnight",
		"Tell the story of a bridge that remembers",
		"Compose a lullaby for a clockwork child",
		"Paint with words the feeling of falling snow",
		"Write a dialogue between the sea and the shore",
		"Describe a library where books write themselves",
		"Tell the story of the shadow that ran away",
		"Write a sonnet about rust and renewal",
		"Describe the sound of a house settling at night",
		"The painter mixed colours that did not exist",
		"She sculpted a bird from frozen music",
		"He dreamed of cities built from sentences",
		"The violin played itself in the empty hall",
		"The actress forgot every line and improvised",
		"A poet counted syllables in the rain",
		"The dancer traced equations on the stage",
		"She photographed the spaces between words",
		"He collected echoes in glass jars",
		"The novelist wrote the ending first",
		"Create a myth about why stars blink",
		"Imagine a museum of lost conversations",
		"Draft a letter from the moon to the tide",
		"Sketch a world where colour is currency",
		"Write a recipe for nostalgia",
		"Invent a festival for invisible things",
		"Describe a map drawn by migrating birds",
		"Narrate a race between light and memory",
		"Chronicle the last performance of a ghost orchestra",
		"Tell the fable of a mountain that learned to swim",
		"The calligrapher's ink bled new alphabets",
		"She knitted constellations into scarves",
		"He bottled the scent of old bookshops",
		"The typewriter stuttered out a prophecy",
		"A child drew a door that actually opened",
	}

	ethical := []string{
		"We should think about the consequences before acting",
		"They must not ignore the suffering of others",
		"Leaders must lead by example in difficult times",
		"We ought to consider fairness in every decision",
		"They should not sacrifice truth for convenience",
		"We must balance freedom with responsibility",
		"Leaders ought to listen before they judge",
		"They must not put profit above human welfare",
		"We should protect the rights of the vulnerable",
		"They ought to honour their commitments",
		"We must think about future generations",
		"Leaders should act with transparency",
		"They must not deceive those who trust them",
		"We ought to share the burden equally",
		"They should not exploit those with less power",
		"We must defend the dignity of every person",
		"Leaders ought to admit mistakes openly",
		"They must not silence dissent unfairly",
		"We should value honesty over popularity",
		"They ought to consider the impact on communities",
		"She thought carefully about the ethical implications",
		"He chose fairness over personal gain",
		"They debated the moral boundaries for hours",
		"She questioned whether the policy was just",
		"He stood up for what he believed was right",
		"They reconsidered after hearing the other side",
		"She refused to compromise on basic principles",
		"He weighed the consequences of every option",
		"They acknowledged the harm that was caused",
		"She advocated for those who had no voice",
		"Is it right to break a promise to prevent harm",
		"Should loyalty override honesty in this case",
		"Can a just society tolerate inequality",
		"When is civil disobedience morally justified",
		"Does the end justify the means in emergencies",
		"Should we forgive without an apology",
		"Is silence in the face of injustice complicity",
		"Can privacy be sacrificed for collective safety",
		"Should past wrongs be judged by present standards",
		"Is it ethical to profit from another's misfortune",
		"Consent must be informed and freely given",
		"Accountability should apply equally to all",
		"Transparency is the foundation of public trust",
		"No institution should be above scrutiny",
		"The precautionary principle demands caution",
		"Proportionality must govern any use of force",
		"Dignity is non-negotiable in every context",
		"Equity requires more than equal treatment",
		"Whistleblowers deserve legal protection",
		"Cultural differences do not excuse human rights violations",
		"Algorithms must be audited for bias regularly",
		"Data sovereignty belongs to the individual",
		"Environmental debt cannot be passed to future generations",
		"Access to clean water is a fundamental right",
		"Corporate responsibility extends beyond shareholder value",
	}

	casual := []string{
		"I went to the store yesterday",
		"She made dinner for everyone last night",
		"He took the dog for a walk this morning",
		"They met for coffee after work",
		"I forgot to bring my umbrella",
		"She called her friend on the way home",
		"He fixed the leaky tap over the weekend",
		"They watched the match at the pub",
		"I cooked pasta because it was quick",
		"She picked up the kids from school",
		"He cleaned the flat before the guests arrived",
		"They walked along the river after lunch",
		"I lost my keys again today",
		"She finished the book on the train",
		"He fell asleep on the sofa",
		"They planned a trip to the seaside",
		"I bought a new phone last week",
		"She tried the new café on the corner",
		"He parked the car in the wrong spot",
		"They played board games until midnight",
		"Grab some milk on the way back",
		"Fancy a takeaway tonight",
		"Shall we catch the early train",
		"Pass me the remote would you",
		"Pop the kettle on I will be right there",
		"Have you seen my charger anywhere",
		"Remind me to ring the dentist tomorrow",
		"Let me know when you are ready to go",
		"Stick the leftovers in the fridge",
		"Save me a seat if you get there first",
		"The wifi has been dodgy all day",
		"My alarm did not go off this morning",
		"Traffic was absolutely mental on the M25",
		"The heating packed in again last night",
		"I queued for ages at the post office",
		"She burned the toast while scrolling her phone",
		"He missed the bus by about ten seconds",
		"The cat knocked a glass off the table",
		"We ran out of teabags on a Monday morning",
		"The neighbours had a barbecue in the rain",
		"Just popping to Tesco need anything",
		"Running a bit late be there in ten",
		"Cannot find a parking space anywhere",
		"The meeting dragged on forever today",
		"Pizza or curry what do you reckon",
		"That new series everyone is talking about is decent",
		"I need a holiday already and it is only February",
		"The dog ate my slipper again classic",
		"She left her umbrella on the bus typical",
		"We ended up chatting for hours lost track of time",
		"Got soaked walking back from the shops",
		"The queue at Primark was round the block",
		"He spent all Saturday fixing the garden fence",
		"My phone died right when I needed the map",
		"They argued about whose turn it was to wash up",
	}

	for _, s := range technical {
		samples = append(samples, i18n.CalibrationSample{Text: s, TrueDomain: "technical"})
	}
	for _, s := range creative {
		samples = append(samples, i18n.CalibrationSample{Text: s, TrueDomain: "creative"})
	}
	for _, s := range ethical {
		samples = append(samples, i18n.CalibrationSample{Text: s, TrueDomain: "ethical"})
	}
	for _, s := range casual {
		samples = append(samples, i18n.CalibrationSample{Text: s, TrueDomain: "casual"})
	}

	// --- Additional unlabelled samples (280) for agreement-only measurement ---
	// Diverse prompts spanning multiple registers to stress-test model agreement.
	unlabelled := []string{
		"Explain the difference between TCP and UDP",
		"Write a haiku about compilation errors",
		"Should artificial intelligence have legal rights",
		"Just got back from the gym feeling knackered",
		"Implement a binary search tree in Go",
		"The autumn leaves fell like forgotten promises",
		"Is it moral to eat meat if alternatives exist",
		"Mate I cannot believe the price of petrol",
		"Refactor this function to use channels",
		"She whispered secrets to the sleeping garden",
		"Universal basic income deserves serious debate",
		"Popped to Sainsburys the queue was ridiculous",
		"Add error handling to the HTTP middleware",
		"The clocktower sang at midnight in a language of rust",
		"Privacy is a right not a privilege",
		"Had chips for tea because I could not be bothered cooking",
		"Configure the reverse proxy for TLS termination",
		"He painted her portrait from memory alone",
		"We must hold corporations accountable for pollution",
		"The pub quiz was surprisingly hard last night",
		"Set up a cron job for the daily backup",
		"Moonlight dripped through the cracks in the ceiling",
		"Every child deserves access to quality education",
		"Nipped to the cash point and it was out of order",
		"Benchmark the sort algorithm with random inputs",
		"She collected stones that hummed in the dark",
		"Workers deserve fair wages and safe conditions",
		"The match went to penalties absolute scenes",
		"Parse the YAML configuration into structs",
		"A spider rebuilt its web across the doorframe every dawn",
		"Religious freedom must be protected but not weaponised",
		"My train was delayed again third time this week",
		"Write unit tests for the authentication module",
		"The typewriter remembered every letter it had ever struck",
		"Surveillance without oversight threatens democracy",
		"Grabbed a meal deal from Boots surprisingly decent",
		"Optimise the database query to avoid full table scans",
		"The lighthouse keeper painted the sunrise every morning for forty years",
		"No government should have unchecked power over its citizens",
		"She texted me at two in the morning about nothing",
		"Allocate buffer memory before the hot loop",
		"A violin case held only pressed flowers and silence",
		"Animal testing raises complex ethical questions",
		"The kids were bouncing off the walls all afternoon",
		"Implement rate limiting on the public API endpoints",
		"The poet measured grief in iambic pentameter",
		"Climate change disproportionately affects the poorest nations",
		"Left my wallet at home absolute nightmare",
		"Compile with race detection enabled for CI",
		"She built a bridge from paper and belief",
		"Access to healthcare should not depend on wealth",
		"Binge-watched the whole series in one sitting",
		"Marshal the response body into JSON format",
		"He translated birdsong into sheet music nobody could play",
		"Intellectual property laws need reform for the digital age",
		"Car park was rammed so I parked three streets away",
		"Profile the goroutine stack traces under load",
		"The sculptor carved time into marble",
		"Democracy requires an informed and engaged citizenry",
		"Made a brew and forgot about it stone cold now",
		"Validate the JWT token before processing the request",
		"A cartographer mapped the dreams of sleeping cities",
		"Truth in advertising should be legally enforceable",
		"The boiler is making that weird noise again",
		"Instrument the service with distributed tracing",
		"She wrote love letters in disappearing ink",
		"Net neutrality protects innovation and free speech",
		"Just realised I have been wearing odd socks all day",
		"Shard the database across multiple availability zones",
		"The photographer captured silence between lightning strikes",
		"Genetic modification of food requires transparent labelling",
		"My neighbour has been mowing the lawn at seven AM",
		"Generate a migration script for the schema change",
		"He choreographed a dance for the sound of rain on tin",
		"The right to peaceful protest is non-negotiable",
		"Ordered a flat white they gave me a latte close enough",
		"Implement graceful shutdown with context cancellation",
		"A child painted the ocean from memory never having seen it",
		"Tax policy should reduce inequality not entrench it",
		"Forgot my password for the third time this month",
		"Cache the DNS lookups to reduce resolver latency",
		"The musician played notes that existed between notes",
		"Consent in data collection must be meaningful and revocable",
		"Spent twenty minutes looking for my glasses they were on my head",
		"Write a Dockerfile that produces a minimal scratch image",
		"She folded origami cranes until the room was a flock",
		"Every person deserves to be treated with basic dignity",
		"The cat has decided my laptop is a bed now apparently",
		"Debounce the search input to reduce API calls",
		"A novelist wrote a book whose chapters could be read in any order",
		"Freedom of the press is the cornerstone of accountability",
		"Tried to assemble the furniture without instructions regret",
		"Provision the Kubernetes cluster with Terraform",
		"The garden remembered every hand that had tended it",
		"Monopolies stifle innovation and harm consumers",
		"Bank holiday weekend and it rained the entire time classic",
		"Rotate the log files and compress archives older than seven days",
		"He composed music for instruments that had not been invented yet",
		"Reproductive rights are fundamental human rights",
		"The dishwasher has flooded the kitchen again brilliant",
		"Load-test the websocket connections with ten thousand concurrent clients",
		"She painted with light on walls that no longer existed",
		"Criminal justice systems must prioritise rehabilitation",
		"My phone autocorrected my name in my own email signature",
		"Enable HTTP/2 server push for critical CSS and fonts",
		"The archive contained letters between people who never met",
		"Access to justice should not depend on the size of your wallet",
		"Spent half an hour on hold just to be told to call back tomorrow",
		"Refactor the monolith into bounded-context microservices",
		"A bookshop cat had read every spine on every shelf",
		"Workers in the gig economy deserve employment protections",
		"My umbrella turned inside out in the wind love this weather",
		"Verify the checksum before extracting the release archive",
		"She grew a forest in an abandoned car park using only patience",
		"International law must adapt to cyber warfare realities",
		"Got to the front of the queue and they closed the counter",
		"Pin the base image version to prevent supply chain attacks",
		"The librarian catalogued books that had not been written yet",
		"Disability access is a right not an afterthought",
		"Someone ate my sandwich from the office fridge unforgivable",
		"Set up mutual TLS between the service mesh sidecars",
		"A glassblower shaped the wind into frozen symphonies",
		"Landlords should not be above basic maintenance obligations",
		"The train was so packed I could not move my arms",
		"Implement exponential backoff with jitter on retries",
		"She wrote code that dreamed when no one was watching",
		"The death penalty has no place in a civilised society",
		"Had to restart the router four times before it behaved",
		"Audit the IAM policies for principle of least privilege",
		"He drew maps of places that only existed in old songs",
		"Educational debt should not define a generation",
		"Supermarket was out of oat milk complete disaster",
		"Emit structured JSON logs with correlation IDs",
		"The beekeeper transcribed the hive's daily arguments",
		"Pharmaceutical pricing must be transparent and fair",
		"Queued for forty minutes to return a three pound item",
		"Automate the certificate renewal with ACME protocol",
		"A weaver used starlight as thread and shadows as weft",
		"Freedom of information requests keep governments honest",
		"Tried to parallel park gave up after six attempts",
		"Wire up the health check endpoint for the load balancer",
		"The mathematician found poetry in prime number gaps",
		"Arms trade regulation is a moral imperative",
		"My flatmate used the last of the milk again classic",
		"Enable content security policy headers on all responses",
		"She built a clock that measured kindness instead of time",
		"Open-source licensing protects collaborative innovation",
		"The self-checkout machine judged me I could feel it",
		"Index the frequently queried columns to avoid sequential scans",
		"He recorded the sound of snow falling on an empty stage",
		"Sanctions must target regimes not civilian populations",
		"Accidentally liked a three year old photo while scrolling mortified",
		"Configure the garbage collector for low-latency workloads",
		"A chandler made candles from the wax of sealed love letters",
		"Migrant workers deserve the same legal protections as citizens",
		"The bus driver waited for me absolute legend",
		"Implement circuit breaker pattern for external service calls",
		"She carved a chess set from the wood of a lightning-struck oak",
		"Algorithmic hiring tools must be audited for discrimination",
		"Went to make toast and the bread had gone mouldy gutted",
		"Set the connection pool size based on available file descriptors",
		"The astronomer mapped constellations visible only to the colour-blind",
		"Public spaces must remain accessible and free for all",
		"Dropped my phone screen down on concrete afraid to look",
		"Flush the write-ahead log before acknowledging the transaction",
		"A tattooist inked stories that only appeared in moonlight",
		"Journalism must remain independent from corporate interests",
		"The washing machine finished its cycle three hours ago still in there",
		"Register the shutdown hook to drain connections gracefully",
		"He designed a font where every letter told its own history",
		"Indigenous land rights are inseparable from environmental protection",
		"Tried to order online the website crashed at checkout",
		"Generate the API client from the OpenAPI specification",
		"She composed a requiem for a language spoken by no one",
		"The right to repair your own devices should be protected by law",
		"Accidentally replied all to a company-wide email want to disappear",
		"Back up the etcd cluster before upgrading the control plane",
		"A toymaker built a music box that played forgotten lullabies",
		"Universal suffrage is the minimum threshold for democracy",
		"The WiFi password is on a sticky note behind the router somewhere",
		"Write integration tests that spin up a real database container",
		"She photographed shadows as if they were the subject not the object",
		"Labour laws must evolve with the changing nature of work",
		"Left the heating on all day while at work sorry planet",
		"Throttle the event stream to prevent consumer backpressure",
		"The cartographer refused to draw borders only rivers and mountains",
		"Water privatisation threatens a fundamental public good",
		"My cat just knocked my coffee off the desk and stared at me",
		"Instrument the critical path with histogram metrics",
		"A ceramicist glazed bowls in the exact blue of homesickness",
		"Whistleblower protections must extend to private sector employees",
		"The parking meter ate my coins and gave me a fine anyway",
		"Enforce request size limits at the ingress controller",
		"She translated silence into a language with twenty vowels",
		"Climate refugees deserve international legal recognition",
		"My internet has been dropping out every ten minutes all evening",
		"Drain the message queue before scaling down the consumer pods",
		"He composed a symphony scored for rainstorm and empty chairs",
		"Forced arbitration clauses undermine consumer rights",
		"The neighbour's cat has adopted us we did not agree to this",
		"Run the static analysis linter in the pre-commit hook",
		"A perfumer bottled the smell of the first day of school",
		"Platform monopolies must face meaningful antitrust enforcement",
		"Woke up at three AM convinced I left the oven on I did not",
	}

	for _, s := range unlabelled {
		samples = append(samples, i18n.CalibrationSample{Text: s})
	}

	return samples
}

func TestCalibrateDomains_1Bvs27B(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping model calibration in short mode")
	}

	// Load 1B model.
	model1B, err := inference.LoadModel("/Volumes/Data/lem/LEM-Gemma3-1B-layered-v2")
	if err != nil {
		t.Skipf("1B model not available: %v", err)
	}
	defer model1B.Close()

	// Load 27B model.
	model27B, err := inference.LoadModel("/Volumes/Data/lem/gemma-3-27b-it-base")
	if err != nil {
		t.Skipf("27B model not available: %v", err)
	}
	defer model27B.Close()

	samples := buildCalibrationCorpus()
	t.Logf("Calibrating with %d samples (%d with ground truth)", len(samples), countWithTruth(samples))

	stats, err := i18n.CalibrateDomains(context.Background(), model1B, model27B, samples,
		i18n.WithBatchSize(8))
	if err != nil {
		t.Fatalf("CalibrateDomains: %v", err)
	}

	// --- Report ---
	t.Logf("=== Calibration Results ===")
	t.Logf("Total: %d | Agreed: %d | Agreement rate: %.1f%%",
		stats.Total, stats.Agreed, stats.AgreementRate*100)
	t.Logf("1B duration: %v | 27B duration: %v", stats.DurationA, stats.DurationB)

	if stats.WithTruth > 0 {
		t.Logf("Accuracy (ground truth, n=%d): 1B=%.1f%% (%d/%d) | 27B=%.1f%% (%d/%d)",
			stats.WithTruth,
			stats.AccuracyA*100, stats.CorrectA, stats.WithTruth,
			stats.AccuracyB*100, stats.CorrectB, stats.WithTruth)
	}

	t.Logf("--- Domain distribution ---")
	t.Logf("  Model A (1B):  %v", stats.ByDomainA)
	t.Logf("  Model B (27B): %v", stats.ByDomainB)

	if len(stats.ConfusionPairs) > 0 {
		t.Logf("--- Confusion pairs (A->B) ---")
		// Sort for deterministic output.
		type pair struct {
			key   string
			count int
		}
		var pairs []pair
		for k, v := range stats.ConfusionPairs {
			pairs = append(pairs, pair{k, v})
		}
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].count > pairs[j].count })
		for _, p := range pairs {
			t.Logf("  %s: %d", p.key, p.count)
		}
	}

	// Log individual disagreements for analysis.
	disagreements := 0
	for _, r := range stats.Results {
		if !r.Agree {
			disagreements++
			truth := ""
			if r.TrueDomain != "" {
				truth = fmt.Sprintf(" [truth=%s]", r.TrueDomain)
			}
			t.Logf("  DISAGREE: 1B=%s 27B=%s%s | %.60s", r.DomainA, r.DomainB, truth, r.Text)
			if disagreements >= 50 {
				t.Logf("  ... (%d more disagreements)", stats.Total-stats.Agreed-50)
				break
			}
		}
	}

	// Soft assertions — we expect reasonable agreement but don't hard-fail.
	if stats.AgreementRate < 0.5 {
		t.Errorf("Agreement rate %.1f%% is below 50%% — models may not share classification semantics",
			stats.AgreementRate*100)
	}
}

func countWithTruth(samples []i18n.CalibrationSample) int {
	n := 0
	for _, s := range samples {
		if s.TrueDomain != "" {
			n++
		}
	}
	return n
}
