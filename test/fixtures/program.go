package fixtures

const (
	ProgramCpp11HelloWorld = `#include <iostream>

using namespace std;

int main() {
	cout << "hello, world!" << endl;
	return 0;
}
`

	ProgramCpp11HelloPerson = `#include <iostream>

using namespace std;

int main() {
	string person;
	cin >> person;
	cout << "hello, " << person << "!" << endl;
	return 0;
}
`

	ProgramCpp11HelloDefaultInteractor = `#include <iostream>

using namespace std;

int main() {
	string s;
	while (cin >> s) {
		cout << 1 << endl;
		cout << "hello, " << s << "!" << endl << flush;
	}
	return 0;
}
`

	ProgramCpp11HelloCustomInteractor = `#include <iostream>
#include <fstream>

using namespace std;

int main() {
	string s;
	while (cin >> s) {
		cout << "hello, " << s << "!" << endl << flush;
	}
	return 0;
}
`
)
