import axios from 'axios';

const instance = axios.create(
    {
        baseURL: 'http://18.219.141.110:3000/api',
        timeout: 15000,
        headers: {
            'Content-Type': 'application/json',
        }
    }
); 

export const parse = async (value) => {
    console.log(value);
    const { data } = await instance.post('/consola', { comando: value });
    return data;
}