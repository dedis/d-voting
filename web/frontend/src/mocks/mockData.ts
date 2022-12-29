import { Results } from 'types/form';

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

const mockRoster: string[] = [
  '123.456.78.9:9000',
  '123.456.78.9:9001',
  '123.456.78.9:9002',
  '123.456.78.9:9003',
  '123.456.78.9:9004',
  '123.456.78.9:9005',
];

const mockForm1: any = {
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
              Hint: '',
            },
          ],
          Ranks: [
            {
              Title: 'Rank the cafeteria',
              ID: (0x19c7).toString(),
              MaxN: 3,
              MinN: 3,
              Choices: ['BC', 'SV', 'Parmentier'],
              Hint: '',
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
          Hint: '',
        },
        {
          Title: 'How did you find the teaching ?',
          ID: (0x41e2).toString(),
          MaxN: 1,
          MinN: 1,
          Choices: ['bad', 'normal', 'good'],
          Hint: '',
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
          Hint: '',
        },
      ],
    },
  ],
};

const mockFormResult11: Results = {
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

const mockFormResult12: Results = {
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

const mockForm2: any = {
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
          Hint: '',
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
          Hint: '',
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
          Hint: '',
        },
      ],

      Ranks: [
        {
          Title: 'Which cafeteria serves the best coffee ?',
          ID: (0xcafe).toString(),
          MaxN: 4,
          MinN: 1,
          Choices: ['Esplanade', 'Giacometti', 'Arcadie', 'Montreux Jazz Cafe'],
          Hint: '',
        },
        {
          Title: 'IN or SC ?',
          ID: (0xbeef).toString(),
          MaxN: 2,
          MinN: 1,
          Choices: ['IN', 'SC'],
          Hint: '',
        },
      ],
      Texts: [],
      Subjects: [],
    },
  ],
};

const mockForm3: any = {
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
          Hint: '',
        },
      ],
      Subjects: [],
    },
  ],
};

const mockFormResult21: Results = {
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

const mockFormResult22: Results = {
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

const mockFormResult23: Results = {
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

const mockFormResult31: Results = {
  SelectResultIDs: [],
  SelectResult: [],
  RankResultIDs: [],
  RankResult: [],
  TextResultIDs: ['PGP'],
  TextResult: [['Alice', 'Pizza', 'Ice cold water', '🍒🍒🍒🍒']],
};

const mockFormResult32: Results = {
  SelectResultIDs: [],
  SelectResult: [],
  RankResultIDs: [],
  RankResult: [],
  TextResultIDs: ['PGP'],
  TextResult: [['Bob', 'Pizza', 'Coke', '🍒🍒🍒']],
};

const mockFormResult33: Results = {
  SelectResultIDs: null,
  SelectResult: null,
  RankResultIDs: null,
  RankResult: null,
  TextResultIDs: null,
  TextResult: null,
};

export {
  mockForm1,
  mockFormResult11,
  mockFormResult12,
  mockForm2,
  mockFormResult21,
  mockFormResult22,
  mockForm3,
  mockFormResult23,
  mockFormResult31,
  mockFormResult32,
  mockFormResult33,
  mockRoster,
  mockNodes,
};
