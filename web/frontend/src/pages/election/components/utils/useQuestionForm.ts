import { useEffect, useState } from 'react';
import { RankQuestion, SelectQuestion, TextQuestion } from 'types/configuration';

const MAX_MINN = 20;

// form hook that handles the form state for all types of questions
const useQuestionForm = (initState: RankQuestion | SelectQuestion | TextQuestion) => {
  const [state, setState] = useState<RankQuestion | SelectQuestion | TextQuestion>(initState);
  const { MinN, Choices } = state;

  // updates the choices length array when minN is greater than the current choices length
  useEffect(() => {
    if (MinN > 0 && MinN < MAX_MINN && Choices.length < MinN) {
      setState({
        ...state,
        Choices: [...Choices, ...new Array(MinN - Choices.length).fill('')],
      });
    }
  }, [MinN, Choices, state]);

  // depending on the type of question, the form state is updated accordingly
  const handleChange = (e) => {
    e.persist();
    switch (e.target.type) {
      case 'number':
        setState({ ...state, [e.target.name]: Number(e.target.value) });
        break;
      case 'text':
        setState({ ...state, [e.target.name]: e.target.value });
        break;
      default:
        break;
    }
  };

  // updates the choices array when the user adds a new choice
  const addChoice: () => void = () => {
    setState({ ...state, Choices: [...state.Choices, ''] });
  };

  // remove a choice from the choices array
  const deleteChoice = (index) => (e) => {
    e.persist();
    if (state.Choices.length > MinN) {
      setState({
        ...state,
        Choices: state.Choices.filter((item: string, idx: number) => idx !== index),
      });
    }
  };

  // update the choice at the given index
  const updateChoice = (index) => (e) => {
    e.persist();
    setState({
      ...state,
      Choices: state.Choices.map((item: string, idx: number) => {
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
