import { useState } from 'react';
import { RankQuestion, SelectQuestion, TextQuestion } from 'types/configuration';

// form hook that handles the form state for all types of questions
const useQuestionForm = (initState: RankQuestion | SelectQuestion | TextQuestion) => {
  const [state, setState] = useState<RankQuestion | SelectQuestion | TextQuestion>(initState);
  const { MinN, Choices } = state;

  // depending on the type of the Exception in the question, the form state is
  // updated accordingly
  const handleChange =
    (Exception?: string, optionnalValues?: number) => (e?: React.ChangeEvent<HTMLInputElement>) => {
      const { value, type, name } = e.target;
      switch (Exception) {
        case 'RankMinMax':
          setState({ ...state, MinN: Number(value), MaxN: Number(value) });
          break;
        case 'addChoiceRank':
          setState({
            ...state,
            Choices: [...Choices, ''],
            MaxN: Choices.length + 1,
            MinN: Choices.length + 1,
          });
          break;
        case 'deleteChoiceRank':
          const filteredChoices = Choices.filter(
            (item: string, idx: number) => idx !== optionnalValues
          );

          setState({
            ...state,
            Choices: filteredChoices,
            MaxN: filteredChoices.length,
            MinN: filteredChoices.length,
          });
          break;
        case 'TextMaxLength':
          if (Number(value) >= 1) {
            setState({ ...state, MaxLength: Number(value) });
          }
          break;
        default:
          e.persist();
          switch (type) {
            case 'number':
              setState({ ...state, [name]: Number(value) });
              break;
            case 'text':
              setState({ ...state, [name]: value });
              break;
            default:
              break;
          }
          break;
      }
    };

  // updates the choices array when the user adds a new choice
  const addChoice = () => {
    setState({ ...state, Choices: [...Choices, ''], MaxN: Choices.length + 1 });
  };

  // remove a choice from the choices array
  const deleteChoice = (index: number) => {
    if (Choices.length > MinN) {
      const filteredChoices = Choices.filter((item: string, idx: number) => idx !== index);

      setState({
        ...state,
        Choices: filteredChoices,
        MaxN: filteredChoices.length,
      });
    }
  };

  // update the choice at the given index
  const updateChoice = (index: number) => (e) => {
    e.persist();
    setState({
      ...state,
      Choices: Choices.map((item: string, idx: number) => {
        if (idx === index) {
          return e.target.value;
        }
        return item;
      }),
    });
  };

  return { state, handleChange, addChoice, deleteChoice, updateChoice };
};

export default useQuestionForm;
