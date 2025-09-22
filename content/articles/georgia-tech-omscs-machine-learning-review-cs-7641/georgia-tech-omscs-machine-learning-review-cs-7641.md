+++
title = "Georgia Tech OMSCS Machine Learning Review | CS 7641"
hook = "Georgia Tech's ML class"
image = "ML2.jpg"
published_at = 2022-08-11T22:12:43-06:00
tags = ["Machine Learning", "OMSCS"]
youtube = "https://youtu.be/Gv1bi24Kzn0"
+++

![](./ChandlerCryingGif-1.gif)
*Summary of this class*

## TL;DR

- Extremely difficult
- 20-25 hours per week on average
- **Enormous** curve, 40% got me a B

## Quick tips

- You don’t need the textbook
    - If you’re interested in the theory go for it, but it doesn’t help much for assignments
- Can work ahead in this class

## What is Machine Learning?

Machine learning is getting computers to make decisions without being explicitly programmed to

## Graded course material

- 2 Exams
    - Closed-notes
    - Each exam is worth 25% of your final grade
    - Difficulty level: Medium, they ask open-ended questions, and all answers can be found in the lectures (watch them for sure)
- 4 Projects
    - Difficulty level: Medium
    - *Tedium* level: Hard
        - I say this because they are very “run these 20 experiments”-esque, which just takes forever to code
    - Writeups: The writeups are very heavily scrutinized, and need to be really well written
        - 45/100 was my average score for writeups
        - <u>Code is not king</u> in this class, writeups are 👑

## Project overview

### Supervised Learning

This assignment in particular has you **use any code package/library** to “implement” the following:

- Decision trees with pruning
- Neural nets
- Boosting
- Support Vector Machines
- k-nearest neighbors

*****PRO TIP***** Use packages like [pytorch](https://pytorch.org/), or [Weka](https://ml.cms.waikato.ac.nz/weka/) to do the heavy lifting for you  
You do not need to code everything from scratch (it even says so in the project spec)

15% of your grade

- Find 2 interesting classification datasets:
    - Can be from anywhere, as long as they are non-trivial and not overly-complex
        - [University of California-Irvine](https://archive.ics.uci.edu/) has some cool public datasets
        - Mine were an MLB baseball statistics dataset, to try and predict the MLB world series winner
        - And an obesity dataset, where you try to guess the obesity level of a person given life characteristics
    - At least 1 should be large
        - Mine had 112 and 2000+ records respectively
    - You will use these in project 3 for:
        - K-means
        - Independent Component Anlaysis (ICA)
        - Insignificant Component Analysis (PCA)


*****PRO TIP***** At least one dataset with more than 2 output classes is best for K-means, ICA and PCA

- Writeup tips:
    - Analyze the <u>behavior</u> of the models
        - Don’t just print the results, really go in-depth here
        - Why did the models behave the way they did?
    - Analyze your datasets in-depth
        - How often do output classes occur?
        - Is there any correlation between variables and output classes?
        - What’s different about each set what are the challenges with each set?
    - Metrics
        - Why did you use the error metric that you used?
        - Were there other error metrics you looked at?

*****PRO TIP***** Use a [fancy paper template](https://www.ieee.org/conferences/publishing/templates) to write your paper, it will get you more points

### Randomized Optimization

This project is very time consuming, because you run a total of 16 experiments

10% of your grade

Here you will experiment with:

- Randomized hill climbing
- Simmulated Annealing
- Genetic Algorithms
- MIMIC

*****PRO TIP***** Watch the lecture videos for this section to understand the algorithms  
They are explained very well in the videos, and MIMIC is Dr. Isabell’s graduate thesis

- Find 3 problems
    - **Bit-string** representation problems are your friend, such as 4-peaks and k-colors
- Going back to project 1
    - You will use:
        - Randomized hill climbing
        - Simulated Annealing
        - Genetic algorithm
    - to train weights for the neural network from project 1 (Supervied Learning)
        - This part was honestly the most confusing, and I suggest doing it first, because it took the longest
- Writeup tips
    - Discuss hyper-parameter tuning
    - Discuss gradient descent performance

### Unsupervised Learning 

Insanely <u>tedious</u>, you do 20 experiments so ❗️start early❗️

- 2 clustering algorithms:
    - k-means clustering
    - Expectation Maximization

- 4 dimensionality algorithms
    - PCA
    - ICA
    - Randomized Projections
    - Any other feature selection algorithm you want

You’ll apply the 2 clustering algorithms to your 2 datasets, then, for each new clustering, you’ll apply all 4 dimensionality reduction algorithms.

So 2 x 2 x 4 = 16 😱

*Then*, you’ll run some stuff on your neural network from project 1 again

So 16 + 4 = **20** experiments total

Then you have to write about all of them 🥲

I have no advice, other than start early

I got 23/100 so good luck!!

### Markov Decision Processes

This one is the most fun in my opinion

Markov Decision Processes is taking stat-action pairs, and training agents to do a task

[Lunar Lander](https://www.gymlibrary.ml/) is one example of an MDP, but you won’t do that in this class

- CS 7642 Reinforcement Learning you do

![](./lunarlander.gif)
*Lunar Lander is an MDP*

- Find 2 MDP problems
    - Not overly complicated
    - One “small” and one “large”
        - I did [Cart-pole](https://www.gymlibrary.ml/environments/classic_control/cart_pole/) and [Taxi](https://www.gymlibrary.ml/environments/toy_text/taxi/)
- Value iteration vs policy iteration
    - To this day I don’t remember the difference
    - But you do both on the problems
- Reinforcement learning algorithm
    - Q-Learning is one popular choice

*****PRO TIP***** If you’re low on steam, I would invest more energy into this project, as it’s 15% of your grade and the most fun in my opinion

- Writeup tips:
    - Graph value and policy iteration convergence side-by-side somehow, either through time spent, iterations spent etc.
    - Do value and policy iteration come to same answer? Why or why not?
    - How do the sizes of the problem affect the algorithms?

## Closing statement

While this class is tough, you can definitely get a B if you stick with it

Don’t worry if you’re getting 30’s/40’s/50’s on assignments, just do your best and it will be okay 🙂
