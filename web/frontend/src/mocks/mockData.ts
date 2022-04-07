const mockElection1 = {
  MainTitle: 'Please give your opinion',
  Scaffold: [
    {
      ID: (0xa2ab).toString(),
      Title: 'Rate the course',
      Order: [(0x3fb2).toString(), (0x41e2).toString(), (0xcd13).toString(), (0xff31).toString()],
      Ranks: [],
      Texts: [
        {
          Title: 'Who were the two best TAs ?',
          ID: (0xcd13).toString(),
          MaxLength: 20,
          Regex: '',
          MaxN: 2,
          MinN: 1,
          Choices: ['TA1', 'TA2'],
        },
      ],
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
      Subjects: [
        {
          Title: "Let's talk about the food",
          ID: (0xff31).toString(),
          Order: [(0xa319).toString(), (0x19c7).toString()],
          Ranks: [
            {
              Title: 'Rank the cafeteria',
              ID: (0x19c7).toString(),
              MaxN: 3,
              MinN: 3,
              Choices: ['BC', 'SV', 'Parmentier'],
            },
          ],
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
          Subjects: [],
        },
      ],
    },
  ],
};

const mockElection2 = {
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
          Choices: [
            'INNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN',
            'SC',
          ],
        },
      ],
      Texts: [],
      Subjects: [],
    },
  ],
};

export { mockElection1, mockElection2 };
