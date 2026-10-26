import { Template } from "./templates";

// communityTemplateFaq answers what people ask about templates users made
// before adding one, for the marketplace pages of listed ones. The sections it
// refers to are on the same page.
export function communityTemplateFaq(template: Template): Template["faq"] {
  const maker =
    template.author.kind === "user" ? template.author.name : "its maker";

  return [
    {
      question: `Who made ${template.name}?`,
      answer: `It was made by ${maker}, a user of Kite, who is responsible for it. Kite reviewed it before listing it in the marketplace, and reviews each new version before apps get it.`,
    },
    {
      question: "Is it safe to add to my server?",
      answer:
        'It runs with your bot and your credits, like commands you build yourself. The "What it can do" section above lists everything that affects your server, like banning members, deleting messages or sending requests to other websites. Check it, and open its commands and event listeners to see each block, before adding it.',
    },
    {
      question: `Can ${maker} see my server or its data?`,
      answer: `No. Adding it copies its commands, event listeners and messages into your app, and ${maker} gets no access to your app or your server. Your secrets and the values of its parameters stay in your app.`,
    },
    {
      question: "How many credits does it use?",
      answer:
        'Adding it is free. Its commands and event listeners use credits when they run, like the ones you build yourself, and "Credit usage" above shows the most each of them costs.',
    },
    {
      question: "Can I change it after adding it?",
      answer:
        "Yes. What it adds are normal commands, event listeners and messages of your app, which you can change in the editor like any other.",
    },
    {
      question: `What happens when ${maker} updates it?`,
      answer: `Kite reviews the new version first. Then your app offers the update and shows what changed. Updating keeps what you changed unless you choose to replace it. If ${maker} deletes the template, what it added stays in your app and keeps working, but gets no more updates.`,
    },
    {
      question: "How do I report it?",
      answer: `Click on Report next to who made it at the top of this page. Kite looks at every report, and ${maker} doesn't see who reported it.`,
    },
  ];
}
