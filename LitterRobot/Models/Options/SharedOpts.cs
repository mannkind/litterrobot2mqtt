using System.Collections.Generic;
using LitterRobot.Models.Shared;
using TwoMQTT.Interfaces;

namespace LitterRobot.Models.Options
{
    /// <summary>
    /// The shared options across the application
    /// </summary>
    public record SharedOpts : ISharedOpts<SlugMapping>
    {
        public const string Section = "LitterRobot";

        /// <summary>
        /// 
        /// </summary>
        /// <typeparam name="SlugMapping"></typeparam>
        /// <returns></returns>
        public List<SlugMapping> Resources { get; init; } = new();
    }
}