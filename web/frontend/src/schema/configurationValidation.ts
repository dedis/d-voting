import * as yup from 'yup';

const idSchema = yup.string().min(1).required();
const titleSchema = yup.string().min(1).required();

const selectsSchema = yup.object({
  ID: yup.lazy(() => idSchema),
  Title: yup.lazy(() => titleSchema),
  MaxN: yup
    .number()
    .min(yup.ref('MinN'), 'Max should be greater than min in selects')
    .integer()
    .required(),
  MinN: yup
    .number()
    .max(yup.ref('MaxN'), `Min should be smaller than max in selects`)
    .min(0)
    .integer()
    .required(),
  Choices: yup
    .array()
    .of(yup.string())
    .min(yup.ref('MinN'), 'Choices array should be at least of length MinN for selects')
    .required(),
});

const ranksSchema = yup.object({
  ID: yup.lazy(() => idSchema),
  Title: yup.lazy(() => titleSchema),
  MaxN: yup
    .number()
    .max(yup.ref('MinN'), 'Max and Min Should be equal in ranks')
    .min(yup.ref('MinN'), 'Max and Min Should be equal in ranks')
    .integer()
    .required(),
  MinN: yup
    .number()
    .max(yup.ref('MaxN'), 'Max and Min Should be equal in ranks')
    .min(yup.ref('MaxN'), 'Max and Min Should be equal in ranks')
    .integer()
    .required(),
  Choices: yup
    .array()
    .of(yup.string())
    .min(yup.ref('MinN'), 'Choices array should be at least of length MinN for ranks')
    .max(yup.ref('MaxN'), 'Choices array length cannot be higher than MaxN for ranks')
    .required(),
});

const textsSchema = yup.object({
  ID: yup.lazy(() => idSchema),
  Title: yup.lazy(() => titleSchema),
  MaxN: yup.number().min(yup.ref('MinN'), 'Max should be greater than min').integer().required(),
  MinN: yup
    .number()
    .max(yup.ref('MaxN'), 'Min should be smaller than max in texts')
    .min(0)
    .integer()
    .required(),
  MaxLength: yup.number().min(1).required(),
  Regex: yup.string(),
  Choices: yup
    .array()
    .of(yup.string())
    .min(yup.ref('MinN'), 'Choices array should be at least of length MinN for texts')
    .max(yup.ref('MaxN'), 'Choices array length cannot be higher than MaxN for texts')
    .required(),
});

const subjectSchema = yup.object({
  ID: yup.lazy(() => idSchema),
  Title: yup.lazy(() => titleSchema),
  Order: yup
    .array()
    .of(yup.string())
    .test({
      name: 'Testing Order',
      message: 'Error Order array is not coherent with the Subject object',
      test() {
        const { Order, Subjects, Ranks, Selects, Texts } = this.parent;
        const onlyUnique = (value, index, self) => self.indexOf(value) === index;

        // If we don't have unique IDs the array is not consistent with the Subject
        if (Order.filter(onlyUnique).length !== Order.length) return false;

        const subjectsID = Subjects.filter(onlyUnique).map((subject) => subject.ID);
        const ranksID = Ranks.filter(onlyUnique).map((rank) => rank.ID);
        const selectsID = Selects.filter(onlyUnique).map((select) => select.ID);
        const textsID = Texts.filter(onlyUnique).map((text) => text.ID);
        const allTypesID = [...subjectsID, ...ranksID, ...selectsID, ...textsID].filter(onlyUnique);

        // Verify that the length of the Order array is exactly the length of the sum of the
        // Subjects, Ranks, Selects and Texts arrays even when duplicates are removed
        if (
          Subjects.length + Ranks.length + Selects.length + Texts.length === Order.length &&
          allTypesID.length === Order.length
        ) {
          let filteredOrder = [...Order];
          for (const id of Order) {
            // If we find the ID in any of the arrays we remove it from our filteredOrder array
            if (
              Subjects.find((subject) => subject.ID === id) ||
              Ranks.find((rank) => rank.ID === id) ||
              Selects.find((select) => select.ID === id) ||
              Texts.find((text) => text.ID === id)
            )
              filteredOrder = filteredOrder.filter((order) => order !== id);
          }
          // If we found all the IDs of our Order in the arrays the test passes
          // meaning that there is exactly one ID for each question type inside the Order array
          if (filteredOrder.length === 0) return true;
          else return false;
        }
        return false;
      },
    })
    .required(),
  Subjects: yup
    .array()
    // @ts-ignore
    .of(yup.lazy(() => subjectSchema))
    .ensure(),
  Selects: yup
    .array()
    // @ts-ignore
    .of(yup.lazy(() => selectsSchema))
    .ensure(),
  Ranks: yup
    .array()
    // @ts-ignore
    .of(yup.lazy(() => ranksSchema))
    .ensure(),
  Texts: yup
    .array()
    // @ts-ignore
    .of(yup.lazy(() => textsSchema))
    .ensure(),
});

const configurationSchema = yup.object({
  MainTitle: yup.lazy(() => titleSchema),
  Scaffold: yup.array().of(subjectSchema).required(),
});

export default configurationSchema;

export { ranksSchema, selectsSchema, subjectSchema, textsSchema };
