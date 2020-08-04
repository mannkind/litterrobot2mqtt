using System.Collections.Generic;
using LitterRobot.Models.Shared;

namespace LitterRobot.Models.Options
{
    /// <summary>
    /// The shared options across the application
    /// </summary>
    public class SharedOpts
    {
        public const string Section = "LitterRobot";

        /// <summary>
        /// 
        /// </summary>
        /// <typeparam name="SlugMapping"></typeparam>
        /// <returns></returns>
        public List<SlugMapping> Resources { get; set; } = new List<SlugMapping>();
    }
}
