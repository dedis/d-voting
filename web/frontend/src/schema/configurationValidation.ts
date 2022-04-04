import * as yup from 'yup';

const idSchema = yup.string().min(1).required();
const titleSchema = yup.string().min(1).required();

const selectsSchema = yup.object({
  ID: yup.lazy(() => idSchema),
  Title: yup.lazy(() => titleSchema),
  MaxN: yup
    .number()
    .test({
      name: 'compare-max-min',
      message: 'Max/Min comparison failed',
      test() {
        const { path, parent } = this;
        const { MaxN, MinN, Choices, ID } = parent;

        if (!Number.isInteger(MaxN)) {
          return this.createError({
            path,
            message: `Max should be an integer in selects, in object ID: ${ID}`,
          });
        }

        if (MaxN < 1) {
          return this.createError({
            path,
            message: `Max should be higher or equal to 1 in selects, in object ID: ${ID}`,
          });
        }

        if (MaxN < MinN) {
          return this.createError({
            path,
            message: `Max should be higher or equal than min in selects, in object ID: ${ID}`,
          });
        }

        if (MaxN >= Choices.length) {
          return this.createError({
            path,
            message: `MaxN should be less or equal to Choices length in selects, in object ID: ${ID}`,
          });
        }
        return true;
      },
    })
    .required(),
  MinN: yup
    .number()
    .max(yup.ref('MaxN'), `Min should be smaller than max in selects`)
    .test({
      name: 'compare-min-max',
      message: 'Max/Min comparison failed',
      test() {
        const { path, parent } = this;
        const { MinN, ID } = parent;

        if (!Number.isInteger(MinN)) {
          return this.createError({
            path,
            message: `Min should be an integer in selects, in object ID: ${ID}`,
          });
        }

        if (MinN < 1) {
          return this.createError({
            path,
            message: `Min should be higher or equal to 1 in selects, in object ID: ${ID}`,
          });
        }
        return true;
      },
    })
    .required(),
  Choices: yup
    .array()
    .of(yup.string())
    .test({
      name: 'compare-choices-min-max',
      message: 'Choice length comparison failed',
      test() {
        const { path, parent } = this;
        const { MaxN, Choices, ID } = parent;

        if (Choices.length < MaxN) {
          return this.createError({
            path,
            message: `Choices array length should be at least equal to Max in selects, in object ID: ${ID}`,
          });
        }
        if (Choices.includes('')) {
          return this.createError({
            path,
            message: `Choices array should not contain empty strings in selects, in object ID: ${ID}`,
          });
        }
        return true;
      },
    })
    .required(),
});

const ranksSchema = yup.object({
  ID: yup.lazy(() => idSchema),
  Title: yup.lazy(() => titleSchema),
  MaxN: yup
    .number()
    .test({
      name: 'compare-max-min',
      message: 'Max/Min comparison failed',
      test() {
        const { path, parent } = this;
        const { MinN, MaxN, ID } = parent;

        if (!Number.isInteger(MaxN)) {
          return this.createError({
            path,
            message: `Max should be an integer in ranks, in object ID: ${ID}`,
          });
        }

        if (MaxN < 2) {
          return this.createError({
            path,
            message: `Max should be higher or equal to 2 in ranks, in object ID: ${ID}`,
          });
        }

        if (MaxN !== MinN) {
          return this.createError({
            path,
            message: `Max and Min Should be equal in ranks, in object ID: ${ID}`,
          });
        }
        return true;
      },
    })
    .required(),
  MinN: yup
    .number()
    .test({
      name: 'compare-min-max',
      message: 'Min/Max comparison failed',
      test() {
        const { path, parent } = this;
        const { MinN, MaxN, ID } = parent;

        if (!Number.isInteger(MinN)) {
          return this.createError({
            path,
            message: `Min should be an integer in ranks, in object ID: ${ID}`,
          });
        }

        if (MinN < 2) {
          return this.createError({
            path,
            message: `Min should be higher or equal to 2 in ranks, in object ID: ${ID}`,
          });
        }

        if (MinN !== MaxN) {
          return this.createError({
            path,
            message: `Min and Max Should be equal in ranks, in object ID: ${ID}`,
          });
        }
        return true;
      },
    })
    .required(),
  Choices: yup
    .array()
    .of(yup.string())
    .test({
      name: 'compare-choices-min-max',
      message: 'Choice length comparison failed',
      test() {
        const { path, parent } = this;
        const { MinN, MaxN, Choices, ID } = parent;

        if (Choices.length !== MaxN || Choices.length !== MinN) {
          return this.createError({
            path,
            message: `Choices array length should be equal to MaxN and MinN in ranks, in object ID: ${ID}`,
          });
        }
        if (Choices.includes('')) {
          return this.createError({
            path,
            message: `Choices array should not contain empty strings in ranks, in object ID: ${ID}`,
          });
        }
        return true;
      },
    })
    .required(),
});

const textsSchema = yup.object({
  ID: yup.lazy(() => idSchema),
  Title: yup.lazy(() => titleSchema),
  MaxN: yup
    .number()
    .test({
      name: 'compare-max-min',
      message: 'Min/Max comparison failed',
      test() {
        const { path, parent } = this;
        const { MinN, MaxN, ID } = parent;

        if (!Number.isInteger(MaxN)) {
          return this.createError({
            path,
            message: `Max should be an integer in texts, in object ID: ${ID}`,
          });
        }

        if (MaxN < 1) {
          return this.createError({
            path,
            message: `Max should be higher or equal to 1 in texts, in object ID: ${ID}`,
          });
        }

        if (MaxN < MinN) {
          return this.createError({
            path,
            message: `Max should be greater than Min in texts, in object ID: ${ID}`,
          });
        }
        return true;
      },
    })
    .required(),
  MinN: yup
    .number()
    .test({
      name: 'compare-min-max',
      message: 'Min/Max comparison failed',
      test() {
        const { path, parent } = this;
        const { MinN, MaxN, ID } = parent;
        if (!Number.isInteger(MinN)) {
          return this.createError({
            path,
            message: `Min should be an integer in texts, in object ID: ${ID}`,
          });
        }
        if (MinN < 1) {
          return this.createError({
            path,
            message: `Min should be higher or equal to 1 in texts, in object ID: ${ID}`,
          });
        }
        if (MinN > MaxN) {
          return this.createError({
            path,
            message: `Min should be smaller than Max in texts, in object ID: ${ID}`,
          });
        }
        return true;
      },
    })
    .required(),
  MaxLength: yup
    .number()
    .test({
      name: 'compare-maxlength',
      message: 'MaxLength value failed',
      test() {
        const { path, parent } = this;
        const { MaxLength, ID } = parent;
        if (!Number.isInteger(MaxLength)) {
          return this.createError({
            path,
            message: `MaxLength should be an integer in texts, in object ID: ${ID}`,
          });
        }
        if (MaxLength < 0) {
          return this.createError({
            path,
            message: `MaxLength should be at least equal to 1 in texts, in object ID: ${ID}`,
          });
        }
        return true;
      },
    })
    .required(),
  Regex: yup.string(),
  Choices: yup
    .array()
    .of(yup.string())
    .test({
      name: 'compare-choices-min-max',
      message: 'Choice length comparison failed',
      test() {
        const { path, parent } = this;
        const { MaxN, Choices, ID } = parent;

        if (Choices.length !== MaxN) {
          return this.createError({
            path,
            message: `Choices array length should be equal to MaxN in texts, in object ID: ${ID}`,
          });
        }
        if (Choices.includes('')) {
          return this.createError({
            path,
            message: `Choices array should not contain empty strings in texts, in object ID: ${ID}`,
          });
        }
        return true;
      },
    })
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
        const { path, parent } = this;
        const { Order, Subjects, Ranks, Selects, Texts } = parent;
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
        }
        return this.createError({
          path,
          message: `Error Order array is not coherent with the Subject in object ID: ${parent.ID}`,
        });
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
