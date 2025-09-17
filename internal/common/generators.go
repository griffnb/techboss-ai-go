package common

import (
	"math/rand"
)

var (
	firstNames = []string{
		"Alice", "Bob", "Charlie", "David", "Eva", "Frank", "Grace", "Helen", "Ivan", "Julia",
		"Kevin", "Linda", "Mark", "Nancy", "Oliver", "Patricia", "Quincy", "Rachel", "Steve", "Tina",
		"Ursula", "Vince", "Wendy", "Xander", "Yvonne", "Zane", "Alex", "Betty", "Chris", "Diana",
		"Edward", "Fiona", "George", "Hannah", "Isaac", "Jane", "Kyle", "Laura", "Mike", "Nina",
		"Oscar", "Pamela", "Quinn", "Rita", "Sam", "Tara", "Ulysses", "Violet", "Walter", "Xena",
		"Yasmine", "Zachary", "Amy", "Ben", "Cathy", "Dan", "Emily", "Felix", "Gina", "Henry",
		"Iris", "Jack", "Katie", "Leo", "Megan", "Noah", "Olivia", "Peter", "Quinn", "Ryan",
		"Sara", "Tom", "Uma", "Victor", "Wendy", "Xander", "Yara", "Zack", "Anna", "Bobby",
		"Carol", "David", "Ella", "Frank", "Grace", "Hank", "Isabel", "Jake", "Kara", "Liam",
		"Mia", "Nate", "Olivia", "Paul", "Rachel", "Sam", "Tina", "Ulysses", "Violet", "Will", "Xena",
		"Yasmine", "Zach", "Ava", "Ben", "Chloe",
	}
	middleNames = []string{
		"Ann", "Brian", "Catherine", "Daniel", "Elaine", "Felix", "Gloria", "Hugo", "Isabel", "John",
		"Katherine", "Louis", "Maria", "Nathan", "Olive", "Paul", "Queenie", "Robert", "Sarah", "Thomas",
		"Ursula", "Victor", "William", "Xander", "Yvonne", "Zane", "Alan", "Barbara", "Chris", "Diana",
		"Ethan", "Fiona", "Greg", "Hannah", "Ivy", "Jack", "Karen", "Leo", "Megan", "Nina",
		"Oliver", "Pamela", "Quincy", "Rita", "Steve", "Tara", "Ulysses", "Violet", "Walter", "Xena",
		"Yasmine", "Zachary", "Alexandra", "Benjamin", "Cathy", "David", "Emily", "Felix", "Gina", "Henry",
		"Iris", "Jacob", "Katie", "Liam", "Mia", "Nate", "Olivia", "Peter", "Quinn", "Ryan",
		"Sara", "Tom", "Uma", "Victor", "Wendy", "Xander", "Yara", "Zack", "Angela", "Brad",
		"Caroline", "Daniel", "Ella", "Frank", "Grace", "Harry", "Isabella", "Jack", "Kara", "Luke",
		"Megan", "Nathan", "Olivia", "Paul", "Rachel", "Sam", "Tina", "Ulysses", "Violet", "Zachary",
		"Alice", "Ben", "Chloe",
	}
	middleInitials = []string{
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J",
		"K", "L", "M", "N", "O", "P", "Q", "R", "S", "T",
		"U", "V", "W", "X", "Y", "Z", "A", "B", "C", "D",
		"E", "F", "G", "H", "I", "J", "K", "L", "M", "N",
		"O", "P", "Q", "R", "S", "T", "U", "V", "W", "X",
		"Y", "Z", "A", "B", "C", "D", "E", "F", "G", "H",
		"I", "J", "K", "L", "M", "N", "O", "P", "Q", "R",
		"S", "T", "U", "V", "W", "X", "Y", "Z", "A", "B",
		"C", "D", "E", "F", "G", "H", "I", "J", "K", "L",
		"M", "N", "O", "P", "Q", "R", "S", "T", "U", "V",
		"W", "X", "Y", "Z",
	}
	lastNames = []string{
		"Smith", "Johnson", "Brown", "Lee", "Wilson", "Davis", "Taylor", "Anderson", "Harris", "Clark",
		"White", "Thomas", "Moore", "Martin", "Hall", "Walker", "Young", "King", "Wright", "Scott",
		"Baker", "Green", "Evans", "Parker", "Adams", "Campbell", "Allen", "Hill", "Roberts", "Turner",
		"Carter", "Mitchell", "Garcia", "Martinez", "Jackson", "Williams", "Jones", "Johnson", "Brown", "Davis",
		"Miller", "Wilson", "Moore", "Taylor", "Anderson", "Thomas", "Walker", "Harris", "Martin", "White",
		"Clark", "King", "Allen", "Scott", "Young", "Hall", "Green", "Baker", "Wright", "Evans",
		"Parker", "Adams", "Carter", "Hill", "Roberts", "Turner", "Garcia", "Martinez", "Jackson", "Williams",
		"Jones", "Johnson", "Brown", "Davis", "Miller", "Wilson", "Moore", "Taylor", "Anderson", "Thomas",
		"Walker", "Harris", "Martin", "White", "Clark", "King", "Allen", "Scott", "Young", "Hall",
		"Green", "Baker", "Wright", "Evans", "Parker", "Adams", "Carter", "Hill", "Roberts", "Turner",
		"Garcia", "Martinez", "Jackson", "Williams",
	}
)

type Name struct {
	FirstName  string `json:"first_name,omitempty"`
	MiddleName string `json:"middle_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
}

func GenerateRandomName() *Name {
	// #nosec
	format := rand.Intn(3)
	// #nosec
	firstName := firstNames[rand.Intn(len(firstNames))]
	// #nosec
	lastName := lastNames[rand.Intn(len(lastNames))]

	switch format {
	case 0: // "first last"
		return &Name{
			FirstName: firstName,
			LastName:  lastName,
		}
	case 1: // "first middle initial last"

		return &Name{
			FirstName: firstName,
			// #nosec
			MiddleName: middleInitials[rand.Intn(len(middleInitials))],
			LastName:   lastName,
		}

	default: // "first middle last"
		return &Name{
			FirstName: firstName,
			// #nosec
			MiddleName: middleNames[rand.Intn(len(middleNames))],
			LastName:   lastName,
		}
	}
}
