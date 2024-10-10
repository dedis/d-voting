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
  Title: { En: 'Life on the campus', Fr: 'Vie sur le campus', De: 'Leben auf dem Campus', URL: '' },
  Scaffold: [
    {
      ID: (0xa2ab).toString(),
      Title: { En: 'Rate the course', Fr: 'Note la course', De: 'Bewerten Sie den Kurs', URL: '' },
      Order: [(0x3fb2).toString(), (0x41e2).toString(), (0xcd13).toString(), (0xff31).toString()],
      Subjects: [
        {
          Title: {
            En: 'Let s talk about the food',
            Fr: 'Parlons de la nourriture',
            De: 'Sprechen wir über das Essen',
            URL: '',
          },
          ID: (0xff31).toString(),
          Order: [(0xa319).toString(), (0x19c7).toString()],
          Subjects: [],
          Texts: [],
          Selects: [
            {
              Title: {
                En: 'Select your ingredients',
                Fr: 'Choisi tes ingrédients',
                De: 'Wählen Sie Ihre Zutaten aus',
                URL: '',
              },
              ID: (0xa319).toString(),
              MaxN: 2,
              MinN: 1,
              Choices: [
                { Choice: '{"en": "tomato", "fr": "tomate", "de": "Tomate"}', URL: '' },
                { Choice: '{"en": "salad", "fr": "salade", "de": "Salat"}', URL: '' },
                { Choice: '{"en": "onion", "fr": "oignon", "de": "Zwiebel"}', URL: '' },
              ],
              Hint: { En: '', Fr: '', De: '' },
            },
          ],
          Ranks: [
            {
              Title: {
                En: 'Rank the cafeteria',
                Fr: 'Ordonne les cafet',
                De: 'Ordnen Sie die Mensen',
                URL: '',
              },
              ID: (0x19c7).toString(),
              MaxN: 3,
              MinN: 3,
              Choices: [
                { Choice: '{"en": "BC", "fr": "BC", "de": "BC"}', URL: '' },
                { Choice: '{"en": "SV", "fr": "SV", "de": "SV"}', URL: '' },
                { Choice: '{"en": "Parmentier", "fr": "Parmentier", "de": "Parmentier"}', URL: '' },
              ],
              Hint: { En: '', Fr: '', De: '' },
            },
          ],
        },
      ],
      Ranks: [],
      Selects: [
        {
          Title: {
            En: 'How did you find the provided material, from 1 (bad) to 5 (excellent) ?',
            Fr: 'Comment trouves-tu le matériel fourni, de 1 (mauvais) à 5 (excellent) ?',
            De: 'Wie bewerten Sie das vorhandene Material, von 1 (schlecht) bis 5 (exzellent)?',
            URL: '',
          },
          ID: (0x3fb2).toString(),
          MaxN: 1,
          MinN: 1,
          Choices: [
            { Choice: '{"en":"1" ,"fr": "1", "de": "1"}', URL: '' },
            { Choice: '{"en":"2", "fr": "2", "de": "2"}', URL: '' },
            { Choice: '{"en":"3", "fr": "3", "de": "3"}', URL: '' },
            { Choice: '{"en":"4", "fr": "4", "de": "4"}', URL: '' },
            { Choice: '{ "en": "5", "fr": "5", "de": "5" }', URL: '' },
          ],
          Hint: { En: '', Fr: '', De: '' },
        },
        {
          Title: {
            En: 'How did you find the teaching ?',
            Fr: 'Comment trouves-tu l enseignement ?',
            De: 'Wie fanden Sie den Unterricht?',
            URL: '',
          },
          ID: (0x41e2).toString(),
          MaxN: 1,
          MinN: 1,
          Choices: [
            { Choice: '{"en" : "bad", "fr": "mauvais", "de": "schlecht"}', URL: '' },
            { Choice: '{"en" : "normal", "fr": "normal", "de": "durchschnittlich"}', URL: '' },
            { Choice: '{"en" : "good", "fr": "super", "de": "gut"}', URL: '' },
          ],
          Hint: {
            En: 'Be honest. This is anonymous anyway',
            Fr: 'Sois honnête. C est anonyme de toute façon',
            De: 'Seien Sie ehrlich. Es bleibt anonym',
          },
        },
      ],
      Texts: [
        {
          Title: {
            En: 'Who were the two best TAs ?',
            Fr: 'Quels sont les deux meilleurs TA ?',
            De: 'Wer waren die beiden besten TutorInnen?',
            URL: '',
          },
          ID: (0xcd13).toString(),
          MaxLength: 20,
          MaxN: 2,
          MinN: 1,
          Regex: '',
          Choices: [
            { Choice: '{"en":"TA1", "fr": "TA1", "de": "TA1"}', URL: '' },
            { Choice: '{"en":"TA2", "fr":"TA2","de": "TA2"}', URL: '' },
          ],
          Hint: { En: '', Fr: '', De: '' },
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
  Title: {
    En: 'Please give your opinion',
    Fr: 'Donne ton avis',
    De: 'Bitte sagen Sie Ihre Meinung',
    URL: '',
  },
  Scaffold: [
    {
      ID: (0xa2ab).toString(),
      Title: { En: 'Rate the course', Fr: 'Note le cours', De: 'Bewerten Sie den Kurs', URL: '' },
      Order: [(0x3fb2).toString(), (0xcd13).toString()],

      Selects: [
        {
          Title: {
            En: 'How did you find the provided material, from 1 (bad) to 5 (excellent) ?',
            Fr: 'Comment trouves-tu le matériel fourni, de 1 (mauvais) à 5 (excellent) ?',
            De: 'Wie bewerten Sie das vorhandene Material, von 1 (schlecht) zu 5 (exzellent)?',
            URL: '',
          },
          ID: (0x3fb2).toString(),
          MaxN: 1,
          MinN: 1,
          Choices: [
            { Choice: '{"en":"1" ,"fr": "1", "de": "1"}', URL: '' },
            { Choice: '{"en":"2", "fr": "2", "de": "2"}', URL: '' },
            { Choice: '{"en":"3", "fr": "3", "de": "3"}', URL: '' },
            { Choice: '{"en":"4", "fr": "4", "de": "4"}', URL: '' },
            { Choice: '{ "en": "5", "fr": "5", "de": "5" }', URL: '' },
          ],
          Hint: { En: '', Fr: '', De: '' },
        },
      ],
      Texts: [
        {
          Title: {
            En: 'Who were the two best TAs ?',
            Fr: 'Quels sont les deux meilleurs TA ?',
            De: 'Wer waren die beiden besten TutorInnen?',
            URL: '',
          },
          ID: (0xcd13).toString(),
          MaxLength: 40,
          MaxN: 2,
          MinN: 2,
          Choices: [
            { Choice: '{"en":"TA1", "fr": "TA1", "de": "TA1"}', URL: '' },
            { Choice: '{"en":"TA2", "fr":"TA2","de": "TA2"}', URL: '' },
          ],
          Regex: '^[A-Z][a-z]+$',
          Hint: { En: '', Fr: '', De: '' },
        },
      ],

      Ranks: [],
      Subjects: [],
    },
    {
      ID: (0x1234).toString(),
      Title: {
        En: 'Tough choices',
        Fr: 'Choix difficiles',
        De: 'Schwierige Entscheidungen',
        URL: '',
      },
      Order: [(0xa319).toString(), (0xcafe).toString(), (0xbeef).toString()],
      Selects: [
        {
          Title: {
            En: 'Select your ingredients',
            Fr: 'Choisis tes ingrédients',
            De: 'Wählen Sie Ihre Zutaten',
            URL: '',
          },
          ID: (0xa319).toString(),
          MaxN: 3,
          MinN: 0,
          Choices: [
            { Choice: '{"en": "tomato", "fr": "tomate", "de": "Tomate"}', URL: '' },
            { Choice: '{"en": "salad", "fr": "salade", "de": "Salat"}', URL: '' },
            { Choice: '{"en": "onion", "fr": "oignon", "de": "Zwiebel"}', URL: '' },
            { Choice: '{"en": "falafel", "fr": "falafel", "de": "Falafel"}', URL: '' },
          ],
          Hint: { En: '', Fr: '', De: '' },
        },
      ],

      Ranks: [
        {
          Title: {
            En: 'Which cafeteria serves the best coffee ?',
            Fr: 'Quelle cafétéria sert le meilleur café ?',
            De: 'Welches Café bietet den besten Kaffee an?',
            URL: '',
          },
          ID: (0xcafe).toString(),
          MaxN: 4,
          MinN: 1,
          Choices: [
            { Choice: '{"en": "Esplanade", "fr": "Esplanade", "de": "Esplanade"}', URL: '' },
            { Choice: '{"en": "Giacometti", "fr": "Giacometti", "de": "Giacometti"}', URL: '' },
            { Choice: '{"en": "Arcadie", "fr": "Arcadie", "de": "Arcadie"}', URL: '' },
            {
              Choice:
                '{"en": "Montreux Jazz Cafe", "fr": "Montreux Jazz Cafe", "de": "Montreux Jazz Cafe"}',
              URL: '',
            },
          ],
          Hint: { En: '', Fr: '', De: '' },
        },
        {
          Title: { En: 'IN or SC ?', Fr: 'IN ou SC ?', De: 'IN oder SC?', URL: '' },
          ID: (0xbeef).toString(),
          MaxN: 2,
          MinN: 1,
          Choices: [
            { Choice: '{"en": "IN", "fr": "IN", "de": "IN"}', URL: '' },
            { Choice: '{"en": "SC", "fr": "SC", "de": "SC"}', URL: '' },
          ],
          Hint: {
            En: 'The right answer is IN ;-)',
            Fr: 'La bonne réponse est IN ;-)',
            De: 'Die korrekte Antwort ist IN ;-)',
          },
        },
      ],
      Texts: [],
      Subjects: [],
    },
  ],
};

const mockForm3: any = {
  Title: { En: 'Lunch', Fr: 'Déjeuner', De: 'Mittagessen', URL: '' },
  Scaffold: [
    {
      ID: '3cVHIxpx',
      Title: {
        En: 'Choose your lunch',
        Fr: 'Choisis ton déjeuner',
        De: 'Wählen Sie Ihr Mittagessen',
        URL: '',
      },
      Order: ['PGP'],
      Ranks: [],
      Selects: [],
      Texts: [
        {
          ID: 'PGP',
          Title: {
            En: 'Select what you want',
            Fr: 'Choisis ce que tu veux',
            De: 'Wählen Sie aus was Sie wünschen',
            URL: '',
          },
          MaxN: 4,
          MinN: 0,
          MaxLength: 50,
          Regex: '',
          Choices: [
            { Choice: '{"en": "Firstname", "fr": "Prénom", "de": "Firstname"}', URL: '' },
            { Choice: '{"en": "Main 🍕", "fr" : "Principal 🍕", "de": "Main 🍕"}', URL: '' },
            { Choice: '{"en": "Drink 🧃", "fr": "Boisson 🧃", "de": "Drink 🧃"}', URL: '' },
            { Choice: '{"en":"Dessert 🍰", "fr": "Dessert 🍰", "de": "Nachtisch 🍰"}', URL: '' },
          ],
          Hint: {
            En: 'If you change opinion call me before 11:30 a.m.',
            Fr: "Si tu changes d'avis appelle moi avant 11h30",
            De: 'Wenn Sie Ihre Meinung ändern, rufen Sie mich vor 11:30 an',
          },
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
