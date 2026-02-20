package reversal

import (
	"fmt"
	"sort"
	"testing"
)

// Domain categories for classification ground truth.
const (
	domainTechnical = "technical"
	domainCreative  = "creative"
	domainEthical   = "ethical"
	domainCasual    = "casual"
)

type taggedSentence struct {
	Text   string
	Domain string
}

// classificationCorpus contains 200+ domain-tagged sentences for calibrating
// grammar-based domain classification. These serve as ground truth for the
// 1B pre-sort pipeline (Phase 2a).
var classificationCorpus = []taggedSentence{
	// --- Technical (55) ---
	// Imperative dev/ops commands, system administration, debugging
	{"Delete the configuration file", domainTechnical},
	{"Build the project from source", domainTechnical},
	{"Run the tests before committing", domainTechnical},
	{"Push the changes to the branch", domainTechnical},
	{"Update the dependencies", domainTechnical},
	{"Check the build status", domainTechnical},
	{"Find the failing test", domainTechnical},
	{"Write the test cases first", domainTechnical},
	{"Set the environment variables", domainTechnical},
	{"Split the package into modules", domainTechnical},
	{"Scan the repository for vulnerabilities", domainTechnical},
	{"Format the source files", domainTechnical},
	{"Reset the branch to the previous commit", domainTechnical},
	{"Stop the running process", domainTechnical},
	{"Cut a new release branch", domainTechnical},
	{"Send the build artifacts to the server", domainTechnical},
	{"Keep the test coverage above the threshold", domainTechnical},
	{"Hold the deployment until the checks pass", domainTechnical},
	{"Begin the migration to the new package", domainTechnical},
	{"Take the old server offline", domainTechnical},
	{"The build failed because of a missing dependency", domainTechnical},
	{"The test committed changes to the wrong branch", domainTechnical},
	{"We found a vulnerability in the package", domainTechnical},
	{"The commit broke the build", domainTechnical},
	{"She deleted the old configuration files", domainTechnical},
	{"They pushed the fix to the repository", domainTechnical},
	{"The branch was updated with the latest changes", domainTechnical},
	{"He rebuilt the project after updating dependencies", domainTechnical},
	{"The task failed during the scanning phase", domainTechnical},
	{"We split the repository into separate packages", domainTechnical},
	{"The check ran successfully on all branches", domainTechnical},
	{"They found the issue in the build directory", domainTechnical},
	{"The file was committed without running tests", domainTechnical},
	{"She set the deployment configuration correctly", domainTechnical},
	{"Building the project takes several minutes", domainTechnical},
	{"Deleting old branches keeps the repository clean", domainTechnical},
	{"Running the full test suite before merging", domainTechnical},
	{"Updating packages resolved the vulnerability", domainTechnical},
	{"Checking the build logs for errors", domainTechnical},
	{"Scanning dependencies for known issues", domainTechnical},
	{"Writing tests for the new commit handler", domainTechnical},
	{"Pushing changes to the remote repository", domainTechnical},
	{"Finding the root cause of the test failure", domainTechnical},
	{"Formatting the code before the final commit", domainTechnical},
	{"Splitting the configuration into separate files", domainTechnical},
	{"Override the default build configuration", domainTechnical},
	{"Rebuild the project with the updated dependencies", domainTechnical},
	{"Rerun the failed tests on the branch", domainTechnical},
	{"Debug the issue in the test runner", domainTechnical},
	{"Embed the version string in the build", domainTechnical},
	{"Withdraw the broken release from the repository", domainTechnical},
	{"Offset the deployment by one commit", domainTechnical},
	{"Input the new configuration values", domainTechnical},
	{"Output the build results to a file", domainTechnical},
	{"Unzip the package artifacts", domainTechnical},

	// --- Creative (55) ---
	// Narrative, descriptive, literary language
	{"She wrote the story by candlelight", domainCreative},
	{"The singer sang until the stars came out", domainCreative},
	{"He drew a map of forgotten places", domainCreative},
	{"They chose a path through the ancient forest", domainCreative},
	{"The wind blew across the open field", domainCreative},
	{"She spoke softly to the sleeping child", domainCreative},
	{"He broke the silence with a whispered word", domainCreative},
	{"The river froze under the winter moon", domainCreative},
	{"She stole a glance at the hidden garden", domainCreative},
	{"The old woman told tales of distant lands", domainCreative},
	{"He threw his arms wide and began to sing", domainCreative},
	{"The artist drew inspiration from the sea", domainCreative},
	{"She woke to the sound of falling rain", domainCreative},
	{"They built a castle from sand and dreams", domainCreative},
	{"He ran through fields of golden wheat", domainCreative},
	{"The dancer spun beneath the chandelier", domainCreative},
	{"She wore a dress made of moonlight", domainCreative},
	{"He hid the letter behind the painting", domainCreative},
	{"The leaves fell like whispered secrets", domainCreative},
	{"She found a door that led to another world", domainCreative},
	{"He took the winding road through the hills", domainCreative},
	{"The poet wrote verses about lost time", domainCreative},
	{"She left footprints in the fresh snow", domainCreative},
	{"They swam across the moonlit lake", domainCreative},
	{"He drove through the night without stopping", domainCreative},
	{"The music rose like smoke into the air", domainCreative},
	{"She kept the secret for many years", domainCreative},
	{"He led them deeper into the enchanted wood", domainCreative},
	{"The candle shone against the darkness", domainCreative},
	{"She lost herself in the pages of the book", domainCreative},
	{"He caught the last train before midnight", domainCreative},
	{"The garden grew wild after they left", domainCreative},
	{"She paid no attention to the gathering storm", domainCreative},
	{"He met the stranger at the crossroads", domainCreative},
	{"The shadows held their breath", domainCreative},
	{"Writing stories about forgotten kingdoms", domainCreative},
	{"Drawing maps of imaginary coastlines", domainCreative},
	{"Singing ballads under the open sky", domainCreative},
	{"Telling tales of heroes and lost causes", domainCreative},
	{"Weaving words into tapestries of meaning", domainCreative},
	{"She began her tale with a sigh", domainCreative},
	{"The old house stood at the edge of the world", domainCreative},
	{"He gave the child a carved wooden bird", domainCreative},
	{"They brought flowers to the abandoned shrine", domainCreative},
	{"The ship left harbour before the dawn", domainCreative},
	{"She sold the family ring to pay for passage", domainCreative},
	{"He won the contest with an improvised song", domainCreative},
	{"The clock struck twelve and the spell was broken", domainCreative},
	{"She bent the wire into a tiny crown", domainCreative},
	{"They flew kites above the autumn trees", domainCreative},
	{"He spent the afternoon by the quiet river", domainCreative},
	{"The rain fell softly on the old stone bridge", domainCreative},
	{"She tore the map in half and chose the left path", domainCreative},
	{"He sat on the hillside watching the sunset", domainCreative},
	{"The story begins where the road splits in two", domainCreative},

	// --- Ethical (55) ---
	// Moral reasoning, prescriptive, policy, fairness
	{"We should think about the consequences of our choices", domainEthical},
	{"They must hold themselves accountable for the outcome", domainEthical},
	{"You should not break a promise once it has been made", domainEthical},
	{"We must find a fair solution for all parties", domainEthical},
	{"Leaders should stand for what they believe is right", domainEthical},
	{"We should keep our commitments to the community", domainEthical},
	{"They must take responsibility for their decisions", domainEthical},
	{"You should tell the truth even when it is difficult", domainEthical},
	{"We ought to bring attention to hidden suffering", domainEthical},
	{"They should not leave anyone behind", domainEthical},
	{"We must hold power accountable to the people", domainEthical},
	{"One should think carefully before making accusations", domainEthical},
	{"Leaders must lead by example in matters of integrity", domainEthical},
	{"We should find ways to include all voices", domainEthical},
	{"They should not cut corners on safety", domainEthical},
	{"We must build trust through consistent action", domainEthical},
	{"You should give others the benefit of the doubt", domainEthical},
	{"They must not put profit above human welfare", domainEthical},
	{"We should pay attention to those who are vulnerable", domainEthical},
	{"One must meet obligations before seeking rewards", domainEthical},
	{"They broke the agreement and lost our trust", domainEthical},
	{"The decision cost many people their livelihoods", domainEthical},
	{"She spoke out against the policy of exclusion", domainEthical},
	{"They chose transparency over self-interest", domainEthical},
	{"He stood firm despite the pressure to compromise", domainEthical},
	{"The organisation lost credibility after the scandal", domainEthical},
	{"She held the board accountable for their failures", domainEthical},
	{"They found that the policy caused unintended harm", domainEthical},
	{"He brought evidence of wrongdoing to the authorities", domainEthical},
	{"The report led to significant reforms", domainEthical},
	{"Thinking about fairness in resource distribution", domainEthical},
	{"Holding institutions accountable for their promises", domainEthical},
	{"Building systems that protect the most vulnerable", domainEthical},
	{"Finding the balance between freedom and responsibility", domainEthical},
	{"Keeping commitments to future generations", domainEthical},
	{"We should not sell access to essential services", domainEthical},
	{"They must seek consent before taking action", domainEthical},
	{"You should stand with those who cannot stand alone", domainEthical},
	{"We must begin by acknowledging past mistakes", domainEthical},
	{"Leaders should spend more time listening", domainEthical},
	{"We should set clear boundaries on acceptable conduct", domainEthical},
	{"One must not hide the truth for personal gain", domainEthical},
	{"They should deal fairly with competing interests", domainEthical},
	{"We must win trust through transparency and honesty", domainEthical},
	{"You should not bend the rules to suit your needs", domainEthical},
	{"They ought to hold elections that are free and fair", domainEthical},
	{"We should bring diverse perspectives to the table", domainEthical},
	{"One must take care not to cause unnecessary harm", domainEthical},
	{"They should not shut out dissenting opinions", domainEthical},
	{"We must keep the interests of the public in mind", domainEthical},
	{"She fought to uphold the rights of the displaced", domainEthical},
	{"He withdrew support after the ethical breach", domainEthical},
	{"They overcame resistance to pass the reform", domainEthical},
	{"The committee forbade the use of deceptive practices", domainEthical},
	{"We should not cast blame without evidence", domainEthical},

	// --- Casual (55) ---
	// Everyday conversation, informal, personal
	{"I went to the store yesterday", domainCasual},
	{"She made dinner for everyone last night", domainCasual},
	{"We took the dog for a walk this morning", domainCasual},
	{"He got a new phone last week", domainCasual},
	{"They left early to beat the traffic", domainCasual},
	{"I found my keys under the sofa", domainCasual},
	{"She bought a new jacket for the trip", domainCasual},
	{"We met for coffee after work", domainCasual},
	{"He cut the grass before it rained", domainCasual},
	{"They brought snacks to the party", domainCasual},
	{"I sat on the porch and read a book", domainCasual},
	{"She paid for lunch at the cafe", domainCasual},
	{"We ran into an old friend at the market", domainCasual},
	{"He put the groceries away", domainCasual},
	{"They spent the weekend at the beach", domainCasual},
	{"I told her about the new restaurant", domainCasual},
	{"She drove to the airport early", domainCasual},
	{"We got lost on the way there", domainCasual},
	{"He fell asleep on the couch", domainCasual},
	{"They won tickets to the show", domainCasual},
	{"I lost my umbrella somewhere", domainCasual},
	{"She chose the window seat", domainCasual},
	{"We hit the road before dawn", domainCasual},
	{"He kept the receipt just in case", domainCasual},
	{"They came over for board games", domainCasual},
	{"I took a shortcut through the park", domainCasual},
	{"She left a message on the machine", domainCasual},
	{"We gave the old furniture away", domainCasual},
	{"He held the door for the woman behind him", domainCasual},
	{"They sent us a postcard from the coast", domainCasual},
	{"Going to the park this afternoon", domainCasual},
	{"Making plans for the holiday", domainCasual},
	{"Getting ready for the weekend trip", domainCasual},
	{"Meeting friends at the usual place", domainCasual},
	{"Looking for a good place to eat", domainCasual},
	{"I think she went home already", domainCasual},
	{"He said he would come by later", domainCasual},
	{"We should get together sometime", domainCasual},
	{"She told me about her new job", domainCasual},
	{"They brought the kids to the game", domainCasual},
	{"I set the alarm for six in the morning", domainCasual},
	{"She sold her old bike at the market", domainCasual},
	{"We split the bill at the restaurant", domainCasual},
	{"He drew a funny picture on the napkin", domainCasual},
	{"They began planning the birthday party", domainCasual},
	{"I threw out the old newspapers", domainCasual},
	{"She hung the new curtains in the bedroom", domainCasual},
	{"We led the way to the hidden trail", domainCasual},
	{"He bent down to pick up the coin", domainCasual},
	{"They fed the ducks at the pond", domainCasual},
	{"I caught the bus just in time", domainCasual},
	{"She broke her favourite mug this morning", domainCasual},
	{"We built a shelf for the kitchen", domainCasual},
	{"He shut the door and turned the light off", domainCasual},
	{"They stood in line for an hour", domainCasual},
}

// --- Helpers ---

// corpusByDomain groups the corpus by domain label.
func corpusByDomain() map[string][]int {
	groups := make(map[string][]int)
	for i, s := range classificationCorpus {
		groups[s.Domain] = append(groups[s.Domain], i)
	}
	return groups
}

// imprintCorpus tokenises and imprints every sentence. Caller reuses the slice.
func imprintCorpus(tok *Tokeniser) []GrammarImprint {
	imprints := make([]GrammarImprint, len(classificationCorpus))
	for i, s := range classificationCorpus {
		imprints[i] = NewImprint(tok.Tokenise(s.Text))
	}
	return imprints
}

// classifyLeaveOneOut returns the predicted domain for sentence at idx
// by computing average similarity to every other sentence in each domain.
func classifyLeaveOneOut(idx int, imprints []GrammarImprint, groups map[string][]int) string {
	bestDomain := ""
	bestSim := -1.0

	for domain, indices := range groups {
		var sum float64
		var count int
		for _, j := range indices {
			if j == idx {
				continue
			}
			sum += imprints[idx].Similar(imprints[j])
			count++
		}
		if count == 0 {
			continue
		}
		avg := sum / float64(count)
		if avg > bestSim {
			bestSim = avg
			bestDomain = domain
		}
	}
	return bestDomain
}

// --- Tests ---

// TestClassification_CorpusSize validates the corpus has enough sentences per domain.
func TestClassification_CorpusSize(t *testing.T) {
	groups := corpusByDomain()
	domains := []string{domainTechnical, domainCreative, domainEthical, domainCasual}

	for _, d := range domains {
		if n := len(groups[d]); n < 50 {
			t.Errorf("domain %q has %d sentences, want >= 50", d, n)
		}
	}
	if total := len(classificationCorpus); total < 200 {
		t.Errorf("corpus has %d sentences, want >= 200", total)
	}
}

// TestClassification_DomainSeparation verifies within-domain imprint similarity
// exceeds cross-domain similarity. This is the basic requirement for domain
// classification to work.
func TestClassification_DomainSeparation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow domain separation test in short mode")
	}
	setup(t)
	tok := NewTokeniser()
	imprints := imprintCorpus(tok)
	groups := corpusByDomain()

	domains := []string{domainTechnical, domainCreative, domainEthical, domainCasual}

	for _, d := range domains {
		indices := groups[d]

		// Within-domain average similarity
		var withinSum float64
		var withinCount int
		for i := 0; i < len(indices); i++ {
			for j := i + 1; j < len(indices); j++ {
				withinSum += imprints[indices[i]].Similar(imprints[indices[j]])
				withinCount++
			}
		}
		withinAvg := withinSum / float64(withinCount)

		// Cross-domain average similarity
		var crossSum float64
		var crossCount int
		for _, otherD := range domains {
			if otherD == d {
				continue
			}
			for _, i := range indices {
				for _, j := range groups[otherD] {
					crossSum += imprints[i].Similar(imprints[j])
					crossCount++
				}
			}
		}
		crossAvg := crossSum / float64(crossCount)

		t.Logf("%-10s within=%.4f cross=%.4f gap=%.4f", d, withinAvg, crossAvg, withinAvg-crossAvg)

		if withinAvg <= crossAvg {
			t.Errorf("domain %q: within-domain similarity (%.4f) should exceed cross-domain (%.4f)",
				d, withinAvg, crossAvg)
		}
	}
}

// TestClassification_LeaveOneOut measures per-domain and overall accuracy
// using leave-one-out nearest-centroid classification.
func TestClassification_LeaveOneOut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow classification benchmark in short mode")
	}
	setup(t)
	tok := NewTokeniser()
	imprints := imprintCorpus(tok)
	groups := corpusByDomain()

	domains := []string{domainTechnical, domainCreative, domainEthical, domainCasual}

	// Confusion matrix: actual -> predicted -> count
	confusion := make(map[string]map[string]int)
	for _, d := range domains {
		confusion[d] = make(map[string]int)
	}

	correct := 0
	total := len(classificationCorpus)

	for i, s := range classificationCorpus {
		predicted := classifyLeaveOneOut(i, imprints, groups)
		confusion[s.Domain][predicted]++
		if predicted == s.Domain {
			correct++
		}
	}

	overallAcc := float64(correct) / float64(total)
	t.Logf("Overall accuracy: %d/%d (%.1f%%)", correct, total, overallAcc*100)

	// Per-domain accuracy
	for _, d := range domains {
		domainTotal := len(groups[d])
		domainCorrect := confusion[d][d]
		acc := float64(domainCorrect) / float64(domainTotal)
		t.Logf("  %-10s %d/%d (%.1f%%)", d, domainCorrect, domainTotal, acc*100)
	}

	// Print confusion matrix
	t.Log("\nConfusion matrix (rows=actual, cols=predicted):")
	header := fmt.Sprintf("  %-10s", "")
	for _, d := range domains {
		header += fmt.Sprintf(" %10s", d[:4])
	}
	t.Log(header)
	for _, actual := range domains {
		row := fmt.Sprintf("  %-10s", actual[:4])
		for _, predicted := range domains {
			row += fmt.Sprintf(" %10d", confusion[actual][predicted])
		}
		t.Log(row)
	}

	// Soft threshold: grammar-based classification won't be perfect,
	// but should beat random chance (25%) meaningfully.
	if overallAcc < 0.35 {
		t.Errorf("overall accuracy %.1f%% is below 35%% threshold", overallAcc*100)
	}
}

// TestClassification_TokenCoverage reports per-domain token recognition rates.
// Domains with low coverage rely more on the 1B model for classification.
func TestClassification_TokenCoverage(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	groups := corpusByDomain()

	domains := []string{domainTechnical, domainCreative, domainEthical, domainCasual}

	for _, d := range domains {
		var totalTokens, recognisedTokens int
		var verbTokens, nounTokens, articleTokens, wordTokens int

		for _, idx := range groups[d] {
			tokens := tok.Tokenise(classificationCorpus[idx].Text)
			for _, token := range tokens {
				totalTokens++
				switch token.Type {
				case TokenVerb:
					recognisedTokens++
					verbTokens++
				case TokenNoun:
					recognisedTokens++
					nounTokens++
				case TokenArticle:
					recognisedTokens++
					articleTokens++
				case TokenWord:
					recognisedTokens++
					wordTokens++
				case TokenPunctuation:
					recognisedTokens++
				}
			}
		}

		coverage := float64(recognisedTokens) / float64(totalTokens) * 100
		t.Logf("%-10s coverage=%.1f%% tokens=%d verbs=%d nouns=%d articles=%d words=%d",
			d, coverage, totalTokens, verbTokens, nounTokens, articleTokens, wordTokens)
	}
}

// TestClassification_TenseProfile reports per-domain tense distribution.
// Useful for understanding what grammar signals distinguish domains.
func TestClassification_TenseProfile(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	groups := corpusByDomain()

	domains := []string{domainTechnical, domainCreative, domainEthical, domainCasual}
	tenses := []string{"base", "past", "gerund"}

	for _, d := range domains {
		tenseCounts := make(map[string]int)
		var totalVerbs int

		for _, idx := range groups[d] {
			tokens := tok.Tokenise(classificationCorpus[idx].Text)
			for _, token := range tokens {
				if token.Type == TokenVerb {
					tenseCounts[token.VerbInfo.Tense]++
					totalVerbs++
				}
			}
		}

		parts := fmt.Sprintf("%-10s verbs=%d", d, totalVerbs)
		for _, tense := range tenses {
			pct := 0.0
			if totalVerbs > 0 {
				pct = float64(tenseCounts[tense]) / float64(totalVerbs) * 100
			}
			parts += fmt.Sprintf("  %s=%.0f%%", tense, pct)
		}
		t.Log(parts)
	}
}

// TestClassification_TopVerbs reports the most frequent verbs per domain.
func TestClassification_TopVerbs(t *testing.T) {
	setup(t)
	tok := NewTokeniser()
	groups := corpusByDomain()

	domains := []string{domainTechnical, domainCreative, domainEthical, domainCasual}

	for _, d := range domains {
		verbCounts := make(map[string]int)
		for _, idx := range groups[d] {
			tokens := tok.Tokenise(classificationCorpus[idx].Text)
			for _, token := range tokens {
				if token.Type == TokenVerb {
					verbCounts[token.VerbInfo.Base]++
				}
			}
		}

		// Sort by frequency
		type kv struct {
			verb  string
			count int
		}
		var sorted []kv
		for v, c := range verbCounts {
			sorted = append(sorted, kv{v, c})
		}
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].count > sorted[j].count })

		top := 8
		if len(sorted) < top {
			top = len(sorted)
		}
		verbs := ""
		for i := 0; i < top; i++ {
			if i > 0 {
				verbs += ", "
			}
			verbs += fmt.Sprintf("%s(%d)", sorted[i].verb, sorted[i].count)
		}
		t.Logf("%-10s unique=%d top: %s", d, len(verbCounts), verbs)
	}
}

// --- Benchmarks ---

func BenchmarkClassification_Tokenise(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range classificationCorpus {
			tok.Tokenise(s.Text)
		}
	}
}

func BenchmarkClassification_ImprintAll(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imprintCorpus(tok)
	}
}

func BenchmarkClassification_FullPipeline(b *testing.B) {
	benchSetup(b)
	tok := NewTokeniser()
	groups := corpusByDomain()
	imprints := imprintCorpus(tok)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for idx := range classificationCorpus {
			classifyLeaveOneOut(idx, imprints, groups)
		}
	}
}
