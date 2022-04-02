import { ChangeEvent, useEffect, useState } from 'react';

const useQuestionForm = (initState: any) => {
  const [state, setState] = useState<any>(initState);
  const { MinN, Choices } = state;

  const handleChange: (e: ChangeEvent<HTMLInputElement>) => void = (e) => {
    e.persist();
    setState({ ...state, [e.target.name]: e.target.value });
  };

  useEffect(() => {
    if (Number(MinN) > 0 && Choices.length < Number(MinN)) {
      setState({
        ...state,
        Choices: [...Choices, ...new Array(Number(MinN) - Choices.length).fill('')],
      });
    }
  }, [MinN, Choices, state]);

  const addChoice: () => void = () => {
    setState({ ...state, Choices: [...state.Choices, ''] });
  };

  const deleteChoice: (index: number) => (e: ChangeEvent<HTMLInputElement>) => void =
    (index) => (e) => {
      e.persist();
      if (state.Choices.length > Number(state.MinN)) {
        setState({
          ...state,
          Choices: state.Choices.filter((item: string, idx: number) => idx !== index),
        });
      }
    };

  const updateChoice: (index: number) => (e: ChangeEvent<HTMLInputElement>) => void =
    (index) => (e) => {
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

  const clearChoices: (e: ChangeEvent<HTMLInputElement>) => void = (e) => {
    e.persist();
    if (Number(state.MinN) > 0) {
      setState({
        ...state,
        Choices: [...new Array(Number(state.MinN)).fill('')],
      });
    } else {
      setState({
        ...state,
        Choices: [],
      });
    }
  };

  return [state, [handleChange, addChoice, deleteChoice, clearChoices, updateChoice]];
};

export default useQuestionForm;
