import * as yup from 'yup';

const idSchema = yup.string().min(1).required();
const titleSchema = yup.object({
  En: yup.string().required(),
  Fr: yup.string(),
  De: yup.string(),
});
const hintSchema = yup.object({
  En: yup.string(),
  Fr: yup.string(),
  De: yup.string(),
});

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
            message: `Max should be an integer in selects [objectID: ${ID}]`,
          });
        }

        if (MaxN < 1) {
          return this.createError({
            path,
            message: `Max should be higher or equal to 1 in selects [objectID: ${ID}]`,
          });
        }

        if (MaxN < MinN) {
          return this.createError({
            path,
            message: `Max should be higher or equal than min in selects [objectID: ${ID}]`,
          });
        }

        if (MaxN > Choices.length) {
          return this.createError({
            path,
            message: `MaxN should be less or equal to Choices length in selects [objectID: ${ID}]`,
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
            message: `Min should be an integer in selects [objectID: ${ID}]`,
          });
        }

        if (MinN < 0) {
          return this.createError({
            path,
            message: `Min should be higher or equal to 0 in selects [objectID: ${ID}]`,
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
            message: `Choices array length should be at least equal to Max in selects [objectID: ${ID}]`,
          });
        }
        if (Choices.includes('')) {
          return this.createError({
            path,
            message: `Choices should not be empty in selects [objectID: ${ID}]`,
          });
        }
        return true;
      },
    })
    .required(),
  Hint: yup.lazy(() => hintSchema),
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
            message: `Max should be an integer in ranks [objectID: ${ID}]`,
          });
        }

        if (MaxN < 2) {
          return this.createError({
            path,
            message: `Max should be higher or equal to 2 in ranks [objectID: ${ID}]`,
          });
        }

        if (MaxN !== MinN) {
          return this.createError({
            path,
            message: `Max and Min Should be equal in ranks [objectID: ${ID}]`,
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
            message: `Min should be an integer in ranks [objectID: ${ID}]`,
          });
        }

        if (MinN < 2) {
          return this.createError({
            path,
            message: `Min should be higher or equal to 2 in ranks [objectID: ${ID}]`,
          });
        }

        if (MinN !== MaxN) {
          return this.createError({
            path,
            message: `Min and Max Should be equal in ranks [objectID: ${ID}]`,
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
            message: `Choices array length should be equal to MaxN and MinN in ranks [objectID: ${ID}]`,
          });
        }
        if (Choices.includes('')) {
          return this.createError({
            path,
            message: `Choices should not be empty in ranks [objectID: ${ID}]`,
          });
        }
        return true;
      },
    })
    .required(),
  Hint: yup.lazy(() => hintSchema),
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
            message: `Max should be an integer in texts [objectID: ${ID}]`,
          });
        }

        if (MaxN < 1) {
          return this.createError({
            path,
            message: `Max should be higher or equal to 1 in texts [objectID: ${ID}]`,
          });
        }

        if (MaxN < MinN) {
          return this.createError({
            path,
            message: `Max should be greater than Min in texts [objectID: ${ID}]`,
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
            message: `Min should be an integer in texts [objectID: ${ID}]`,
          });
        }
        if (MinN < 0) {
          return this.createError({
            path,
            message: `Min should be higher or equal to 0 in texts [objectID: ${ID}]`,
          });
        }
        if (MinN > MaxN) {
          return this.createError({
            path,
            message: `Min should be smaller than Max in texts [objectID: ${ID}]`,
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
            message: `MaxLength should be an integer in texts [objectID: ${ID}]`,
          });
        }
        if (MaxLength <= 0) {
          return this.createError({
            path,
            message: `MaxLength should be at least equal to 1 in texts [objectID: ${ID}]`,
          });
        }
        if (MaxLength > 1000) {
          return this.createError({
            path,
            message: `MaxLength should not exceed 1000 in texts [objectID: ${ID}]`,
          });
        }
        return true;
      },
    })
    .required(),
  Regex: yup.string().min(0),
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
            message: `Choices array length should be equal to the number of choices [objectID: ${ID}]`,
          });
        }
        return true;
      },
    })
    .required(),
  Hint: yup.lazy(() => hintSchema),
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
        const { Order, Subjects, Ranks, Selects, Texts, ID } = parent;
        const onlyUnique = (value, index, self) => self.indexOf(value) === index;

        // If we don't have unique IDs the array is not consistent with the Subject
        if (Order.filter(onlyUnique).length !== Order.length) return false;

        const subjectsID = Subjects.filter(onlyUnique).map((subject) => subject.ID);
        const ranksID = Ranks.filter(onlyUnique).map((rank) => rank.ID);
        const selectsID = Selects.filter(onlyUnique).map((select) => select.ID);
        const textsID = Texts.filter(onlyUnique).map((text) => text.ID);
        const allTypesUniqueID = [...subjectsID, ...ranksID, ...selectsID, ...textsID].filter(
          onlyUnique
        );

        // Verify that the length of the Order array is exactly the length of the sum of the
        // Subjects, Ranks, Selects and Texts arrays.
        if (Subjects.length + Ranks.length + Selects.length + Texts.length !== Order.length) {
          return this.createError({
            path,
            message: `Order array length is incoherent with the other fields [objectID: ${ID}]`,
          });
        }

        // Verify that we only have unique IDs in the Order array
        if (allTypesUniqueID.length !== Order.length) {
          return this.createError({
            path,
            message: `Order array has duplicate IDs [objectID: ${ID}]`,
          });
        }

        // check if the ID corresponds to any subjects or question object, otherwise return an error
        for (const id of Order) {
          if (
            !Subjects.find((subject) => subject.ID === id) &&
            !Ranks.find((rank) => rank.ID === id) &&
            !Selects.find((select) => select.ID === id) &&
            !Texts.find((text) => text.ID === id)
          ) {
            return this.createError({
              path,
              message: `The ID: ${id} doesn't match any of the subjects or question object [objectID: ${ID}]`,
            });
          }
        }
        return true;
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
  Title: yup.lazy(() => titleSchema),
  Scaffold: yup.array().of(subjectSchema).required(),
});

export default configurationSchema;

export { ranksSchema, selectsSchema, subjectSchema, textsSchema };
