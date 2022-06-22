import { Results } from 'types/election';

const mockNodes: string[] = [
  '123.456.78.9:9000',
  '123.456.78.9:9001',
  '123.456.78.9:9002',
  '123.456.78.9:9003',
  '123.456.78.9:9004',
  '123.456.78.9:9005',
  '123.456.78.9:9006',
  '123.456.78.9:9007',
  '123.456.78.9:9008',
  '123.456.78.9:9009',
];

const mockRoster: string[] = ['123.456.78.9:9000', '123.456.78.9:9001', '123.456.78.9:9002'];

const mockElection1: any = {
  MainTitle: 'Life on the campus',
  Scaffold: [
    {
      ID: (0xa2ab).toString(),
      Title: 'Rate the course',
      Order: [(0x3fb2).toString(), (0x41e2).toString(), (0xcd13).toString(), (0xff31).toString()],
      Subjects: [
        {
          Title: "Let's talk about the food",
          ID: (0xff31).toString(),
          Order: [(0xa319).toString(), (0x19c7).toString()],
          Subjects: [],
          Texts: [],
          Selects: [
            {
              Title: 'Select your ingredients',
              ID: (0xa319).toString(),
              MaxN: 2,
              MinN: 1,
              Choices: ['tomato', 'salad', 'onion'],
            },
          ],
          Ranks: [
            {
              Title: 'Rank the cafeteria',
              ID: (0x19c7).toString(),
              MaxN: 3,
              MinN: 3,
              Choices: ['BC', 'SV', 'Parmentier'],
            },
          ],
        },
      ],
      Ranks: [],
      Selects: [
        {
          Title: 'How did you find the provided material, from 1 (bad) to 5 (excellent) ?',
          ID: (0x3fb2).toString(),
          MaxN: 1,
          MinN: 1,
          Choices: ['1', '2', '3', '4', '5'],
        },
        {
          Title: 'How did you find the teaching ?',
          ID: (0x41e2).toString(),
          MaxN: 1,
          MinN: 1,
          Choices: ['bad', 'normal', 'good'],
        },
      ],
      Texts: [
        {
          Title: 'Who were the two best TAs ?',
          ID: (0xcd13).toString(),
          MaxLength: 20,
          MaxN: 2,
          MinN: 1,
          Regex: '',
          Choices: ['TA1', 'TA2'],
        },
      ],
    },
  ],
};

const mockElectionResult11: Results = {
  SelectResultIDs: [(0x3fb2).toString(), (0x41e2).toString(), (0xa319).toString()],
  SelectResult: [
    [true, false, false, false, false],
    [true, false, false],
    [false, true, true],
  ],
  RankResultIDs: [(0x19c7).toString()],
  RankResult: [[0, 1, 2]],
  TextResultIDs: [(0xcd13).toString()],
  TextResult: [['Noémien', 'Pierluca']],
};

const mockElectionResult12: Results = {
  SelectResultIDs: [(0x3fb2).toString(), (0x41e2).toString(), (0xa319).toString()],
  SelectResult: [
    [false, false, false, true, false],
    [false, false, true],
    [true, false, true],
  ],
  RankResultIDs: [(0x19c7).toString()],
  RankResult: [[0, 2, 1]],
  TextResultIDs: [(0xcd13).toString()],
  TextResult: [['Noémien', 'Pierluca']],
};

const mockElection2: any = {
  MainTitle: 'Please give your opinion',
  Scaffold: [
    {
      ID: (0xa2ab).toString(),
      Title: 'Rate the course',
      Order: [(0x3fb2).toString(), (0xcd13).toString()],

      Selects: [
        {
          Title: 'How did you find the provided material, from 1 (bad) to 5 (excellent) ?',
          ID: (0x3fb2).toString(),
          MaxN: 1,
          MinN: 1,
          Choices: ['1', '2', '3', '4', '5'],
        },
      ],
      Texts: [
        {
          Title: 'Who were the two best TAs ?',
          ID: (0xcd13).toString(),
          MaxLength: 40,
          MaxN: 2,
          MinN: 2,
          Choices: ['TA1', 'TA2'],
          Regex: '^[A-Z][a-z]+$',
        },
      ],

      Ranks: [],
      Subjects: [],
    },
    {
      ID: (0x1234).toString(),
      Title: 'Tough choices',
      Order: [(0xa319).toString(), (0xcafe).toString(), (0xbeef).toString()],
      Selects: [
        {
          Title: 'Select your ingredients',
          ID: (0xa319).toString(),
          MaxN: 3,
          MinN: 0,
          Choices: ['tomato', 'salad', 'onion', 'falafel'],
        },
      ],

      Ranks: [
        {
          Title: 'Which cafeteria serves the best coffee ?',
          ID: (0xcafe).toString(),
          MaxN: 4,
          MinN: 1,
          Choices: ['Esplanade', 'Giacometti', 'Arcadie', 'Montreux Jazz Cafe'],
        },
        {
          Title: 'IN or SC ?',
          ID: (0xbeef).toString(),
          MaxN: 2,
          MinN: 1,
          Choices: ['IN', 'SC'],
        },
      ],
      Texts: [],
      Subjects: [],
    },
  ],
};

const mockElection3: any = {
  MainTitle: 'Lunch',
  Scaffold: [
    {
      ID: '3cVHIxpx',
      Title: 'Choose your lunch',
      Order: ['PGP'],
      Ranks: [],
      Selects: [],
      Texts: [
        {
          ID: 'PGP',
          Title: 'Select what you want',
          MaxN: 4,
          MinN: 0,
          MaxLength: 50,
          Regex: '',
          Choices: ['Firstname', 'Main 🍕', 'Drink 🧃', 'Dessert 🍰'],
        },
      ],
      Subjects: [],
    },
  ],
};

const mockElectionResult21: Results = {
  SelectResultIDs: [(0x3fb2).toString(), (0xa319).toString()],
  SelectResult: [
    [true, false, false, false, false],
    [false, true, true, false],
  ],
  RankResultIDs: [(0xcafe).toString(), (0xbeef).toString()],
  RankResult: [
    [2, 3, 1, 0],
    [0, 1],
  ],
  TextResultIDs: [(0xcd13).toString()],
  TextResult: [['Jane Doe', 'John Smith']],
};

const mockElectionResult22: Results = {
  SelectResultIDs: [(0x3fb2).toString(), (0xa319).toString()],
  SelectResult: [
    [false, false, true, false, false],
    [true, true, true, false],
  ],
  RankResultIDs: [(0xcafe).toString(), (0xbeef).toString()],
  RankResult: [
    [3, 0, 1, 2],
    [1, 0],
  ],
  TextResultIDs: [(0xcd13).toString()],
  TextResult: [['Jane Doe', 'John Smith']],
};

const mockElectionResult23: Results = {
  SelectResultIDs: [(0x3fb2).toString(), (0xa319).toString()],
  SelectResult: [
    [false, false, false, true, false],
    [false, true, false, true],
  ],
  RankResultIDs: [(0xcafe).toString(), (0xbeef).toString()],
  RankResult: [
    [3, 0, 1, 2],
    [1, 0],
  ],
  TextResultIDs: [(0xcd13).toString()],
  TextResult: [['Another Name', 'Jane Doe']],
};

const mockElectionResult31: Results = {
  SelectResultIDs: [],
  SelectResult: [],
  RankResultIDs: [],
  RankResult: [],
  TextResultIDs: ['PGP'],
  TextResult: [['Alice', 'Pizza', 'Ice cold water', '🍒🍒🍒🍒']],
};

const mockElectionResult32: Results = {
  SelectResultIDs: [],
  SelectResult: [],
  RankResultIDs: [],
  RankResult: [],
  TextResultIDs: ['PGP'],
  TextResult: [['Bob', 'Pizza', 'Coke', '🍒🍒🍒']],
};

const mockElectionResult33: Results = {
  SelectResultIDs: null,
  SelectResult: null,
  RankResultIDs: null,
  RankResult: null,
  TextResultIDs: null,
  TextResult: null,
};

export {
  mockElection1,
  mockElectionResult11,
  mockElectionResult12,
  mockElection2,
  mockElectionResult21,
  mockElectionResult22,
  mockElection3,
  mockElectionResult23,
  mockElectionResult31,
  mockElectionResult32,
  mockElectionResult33,
  mockRoster,
  mockNodes,
};
