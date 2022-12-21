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
  MainTitle:
    '{ "en" : "Life on the campus", "fr" : "Vie sur le campus", "de" : "Vie sur le campus"}',
  Scaffold: [
    {
      ID: (0xa2ab).toString(),
      Title: '{ "en" : "Rate the course", "fr" : "Note la course", "de" : "Rate the course"}',
      Order: [(0x3fb2).toString(), (0x41e2).toString(), (0xcd13).toString(), (0xff31).toString()],
      Subjects: [
        {
          Title: 
           '{ "en" : "Let s talk about the food", "fr" : "Parlons de la nourriture", "de" : "Let s talk about food"}',
          ID: (0xff31).toString(),
          Order: [(0xa319).toString(), (0x19c7).toString()],
          Subjects: [],
          Texts: [],
          Selects: [
            {
              Title:
               '{ "en" : "Select your ingredients", "fr" : "Choisi tes ingrédients", "de" : "Select your ingredients"}',
              ID: (0xa319).toString(),
              MaxN: 2,
              MinN: 1,
              Choices: ['{"en": "tomato", "fr": "tomate", "de": "tomato"}','{"en": "salad", "fr": "salade", "de": "salad"}', '{"en": "onion", "fr": "oignon", "de": "onion"}'],
            },
          ],
          Ranks: [
            {
              Title: '{ "en" : "Rank the cafeteria", "fr" : "Ordonne les cafet", "de" : "Rank the cafeteria"}',
              ID: (0x19c7).toString(),
              MaxN: 3,
              MinN: 3,
              Choices: ['{"en": "BC", "fr": "BC", "de": "BC"}', '{"en": "SV", "fr": "SV", "de": "SV"}','{"en": "Parmentier", "fr": "Parmentier", "de": "Parmentier"}'],
            },
          ],
        },
      ],
      Ranks: [],
      Selects: [
        {
          Title: '{"en" : "How did you find the provided material, from 1 (bad) to 5 (excellent) ?", "fr" : "Comment trouves-tu le matériel fourni, de 1 (mauvais) à 5 (excellent) ?", "de" : "How did you find the provided material, from 1 (bad) to 5 (excellent) ?"}',
          ID: (0x3fb2).toString(),
          MaxN: 1,
          MinN: 1,
          Choices: ['{"en":"1" ,"fr": "1", "de": "1"}', '{"en":"2", "fr": "2", "de": "2"}', '{"en":"3", "fr": "3", "de": "3"}', '{"en":"4", "fr": "4", "de": "4"}', {'en':'5', 'fr': '5', 'de': '5'}],
        },
        {
          Title: '{"en": "How did you find the teaching ?","fr": "Comment trouves-tu l enseignement ?","de": "How did you find the teaching ?"}',
          ID: (0x41e2).toString(),
          MaxN: 1,
          MinN: 1,
          Choices: ['{"en" : "bad", "fr": "mauvais", "de": "bad"}','{"en" : "normal", "fr": "normal", "de": "normal"}', '{"en" : "good", "fr": "super", "de": "good"}'],
        },
      ],
      Texts: [
        {
          Title: '{ "en" : Who were the two best TAs ?, "fr" : "Quels sont les deux meilleurs TA ? "de" : Who were the two best TAs ?} ',
          ID: (0xcd13).toString(),
          MaxLength: 20,
          MaxN: 2,
          MinN: 1,
          Regex: '',
          Choices: ['{"en":"TA1", "fr": "TA1", "de": "TA1"}','{"en":"TA2", "fr":"TA2","de:"TA2"}'],
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
  MainTitle: '{"en": "Please give your opinion", "fr": "Donne ton avis", "de": "Please give your opinion"}',
  Scaffold: [
    {
      ID: (0xa2ab).toString(),
      Title: '{"en": "Rate the course", "fr": "Note le cours", "de": "Rate the course"}',
      Order: [(0x3fb2).toString(), (0xcd13).toString()],

      Selects: [
        {
          Title: '{"How did you find the provided material, from 1 (bad) to 5 (excellent) ?", "fr" : "Comment trouves-tu le matériel fourni, de 1 (mauvais) à 5 (excellent) ?", "de" : "How did you find the provided material, from 1 (bad) to 5 (excellent) ?"}',
          ID: (0x3fb2).toString(),
          MaxN: 1,
          MinN: 1,
          Choices:['{"en":"1" ,"fr": "1", "de": "1"}', '{"en":"2", "fr": "2", "de": "2"}', '{"en":"3", "fr": "3", "de": "3"}', '{"en":"4", "fr": "4", "de": "4"}', {'en':'5', 'fr': '5', 'de': '5'}],
        },
      ],
      Texts: [
        {
          Title: 'Who were the two best TAs ?',
          ID: (0xcd13).toString(),
          MaxLength: 40,
          MaxN: 2,
          MinN: 2,
          Choices: ['{"en":"TA1", "fr": "TA1", "de": "TA1"}','{"en":"TA2", "fr":"TA2","de:"TA2"}'],
          Regex: '^[A-Z][a-z]+$',
        },
      ],

      Ranks: [],
      Subjects: [],
    },
    {
      ID: (0x1234).toString(),
      Title: '{"Tough choices", "fr": "Choix difficiles", "de": "Tough choices"}',
      Order: [(0xa319).toString(), (0xcafe).toString(), (0xbeef).toString()],
      Selects: [
        {
          Title: '{"Select your ingredients", "fr": "Choisis tes ingrédients", "de": "Select your ingredients"}',
          ID: (0xa319).toString(),
          MaxN: 3,
          MinN: 0,
          Choices: ['{"en": "tomato", "fr": "tomate", "de": "tomato"}','{"en": "salad", "fr": "salade", "de": "salad"}', '{"en": "onion", "fr": "oignon", "de": "onion"}','{"en": "falafel, "fr": "falafel", "de": "falafel"}'],
        },
      ],

      Ranks: [
        {
          Title: '{"en": "Which cafeteria serves the best coffee ?", "fr": "Quelle cafétéria sert le meilleur café ?", "de": "Which cafeteria serves the best coffee ?"}',
          ID: (0xcafe).toString(),
          MaxN: 4,
          MinN: 1,
          Choices: ['{"en": "Esplanade", "fr": "Esplanade", "de": "Esplanade"}','{"en": "Giacometti", "fr": "Giacometti", "de": "Giacometti"}','{"en": "Arcadie", "fr": "Arcadie", "de": "Arcadie"}','{"en": "Montreux Jazz Cafe", "fr": "Montreux Jazz Cafe", "de": "Montreux Jazz Cafe"}'],
        },
        {
          Title: '{"en": "IN or SC ?", "fr": "IN ou SC ?", "de": "IN or SC ?"}',
          ID: (0xbeef).toString(),
          MaxN: 2,
          MinN: 1,
          Choices: ['{"en": "IN", "fr": "IN", "de": "IN"}','{"en": "SC", "fr": "SC", "de": "SC"}'],
        },
      ],
      Texts: [],
      Subjects: [],
    },
  ],
};

const mockForm3: any = {
  MainTitle: '{"en": "Lunch", "fr": "Déjeuner", "de": "Lunch"}',
  Scaffold: [
    {
      ID: '3cVHIxpx',
      Title: '{"en": "Choose your lunch", "fr": "Choisis ton déjeuner", "de": "Choose your lunch"}',
      Order: ['PGP'],
      Ranks: [],
      Selects: [],
      Texts: [
        {
          ID: 'PGP',
          Title: '{"en": "Select what you want", "fr": "Choisis ce que tu veux", "de": "Select what you want"}',
          MaxN: 4,
          MinN: 0,
          MaxLength: 50,
          Regex: '',
          Choices: ['{"en": "Firstname", "fr": "Prénom", "de": "Firstname"}','{"en": "Main 🍕", "fr" : "Principal 🍕", "de": "Main 🍕"}', '{"en": "Drink 🧃", "fr": "Boisson 🧃", "de": "Drink 🧃"}','{"en":"Dessert 🍰", "fr": "Dessert 🍰", "de": "Dessert 🍰"}'],
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
